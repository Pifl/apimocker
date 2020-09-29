package mock

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

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
	Responses []*response
}

type response struct {
	Body    string
	Code    code
	Headers []header
	Delay   delay
}

type delay struct {
	Value    string
	duration time.Duration
}

type code struct {
	Value int
}

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

func (e *delay) UnmarshalJSON(b []byte) error {
	var tmp string
	json.Unmarshal(b, &tmp)
	duration, err := time.ParseDuration(tmp)
	if err != nil {
		duration = 0
	}
	delay := delay{
		Value:    tmp,
		duration: duration,
	}
	*e = delay
	return nil
}
func (e delay) MarshalJSON() ([]byte, error) {
	delay := e.Value
	b, err := json.Marshal(delay)
	return b, err
}

func (c *code) UnmarshalJSON(b []byte) error {
	var tmp int
	json.Unmarshal(b, &tmp)
	code := code{
		Value: tmp,
	}
	*c = code
	return nil
}
func (c code) MarshalJSON() ([]byte, error) {
	code := c.Value
	b, err := json.Marshal(code)
	return b, err
}

func New(b []byte) (*Mock, error) {
	var mock Mock
	err := json.Unmarshal(b, &mock)

	//Validate Logic / Set defaults
	return &mock, err
}

func (m *Mock) next() *response {
	if len(m.Responses) == 0 {
		// Set 404
		var rsp response = response{
			Body: "Not Found",
		}
		return &rsp
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
		start := time.Now()

		rsp := m.next()

		for _, h := range rsp.Headers {
			w.Header().Set(h.Name, h.Value)
		}
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(rsp.Code.Value)

		desired := rsp.Delay.duration
		end := time.Now()
		elapsed := end.Sub(start)
		delay := desired - elapsed
		time.Sleep(delay)

		fmt.Fprint(w, rsp.Body+"\n")

	})
}
