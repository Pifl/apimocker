package mock

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// Mock is the custom type for a mock
type Mock struct {
	ID   int
	Port int

	Name     string
	Path     string
	Selector selector

	index     int
	Responses []response
}

type response struct {
	Body    string
	Code    code
	Headers []header
}

type code int

type header struct {
	Name  string
	Value string
}

type selector string

const (
	sequence selector = "Sequence"
	random            = "Random"
	script            = "Script"
)

func (s *selector) UnmarshalJSON(b []byte) error {
	var tmp string
	json.Unmarshal(b, &tmp)
	tmp = strings.Title(strings.ToLower(tmp))
	selectorT := selector(tmp)
	switch selectorT {
	case sequence, random, script:
		*s = selectorT
		return nil
	}
	return errors.New("Invalid Selector type")
}

func New(b []byte) (*Mock, error) {
	var mock Mock
	err := json.Unmarshal(b, &mock)
	//Validate Logic / Set defaults
	return &mock, err
}

func (m *Mock) next() response {
	if len(m.Responses) == 0 {
		// Set 404
		var rsp response = response{
			Body: "Not Found",
		}
		return rsp
	}
	i := m.index
	switch m.Selector {
	case sequence:
		i++
		if i >= len(m.Responses) {
			i = 0
		}
	case random:
		i = rand.Intn(len(m.Responses))
	default:
		i = 0
	}
	m.index = i
	return m.Responses[m.index]
}

// Handler adds a mock specifc handler to the router of the host server
func (m *Mock) Handler(router *httprouter.Router) {
	router.GET(m.Path, func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		fmt.Fprint(w, m.next().Body+"\n")
	})
}
