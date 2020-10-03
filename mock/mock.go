package mock

import (
	"crypto/md5"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/julienschmidt/httprouter"
)

// Mock is the custom type for a mock
type Mock struct {
	ID string

	Name     string
	Path     string
	Selector selector

	index       int
	environment map[string]interface{}
	program     *vm.Program
	Responses   []*response

	Instances int
}

type response struct {
	Body    body
	Code    code
	Headers []header
	Delay   delay
}

type body struct {
	Encoding string
	Content  []byte
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

// UnmarshalJSON for the Response bodies first treats it as two string then based on the encoding
// converts the contents to a []byte
func (b *body) UnmarshalJSON(bytes []byte) error {

	type intermediate struct {
		Encoding string
		Content  string
	}
	var intermediary intermediate
	json.Unmarshal(bytes, &intermediary)

	var data []byte
	var err error
	// Check if encoding is one of the accepted values
	switch intermediary.Encoding {
	case "base64":
		data, err = base64.StdEncoding.DecodeString(intermediary.Content)
		if err != nil {
			fmt.Println("error:", err)
			return err
		}
	case "base32":
		data, err = base32.StdEncoding.DecodeString(intermediary.Content)
		if err != nil {
			fmt.Println("error:", err)
			return err
		}
	default:
		intermediary.Encoding = "raw"
		data = []byte(intermediary.Content)
	}

	newbody := body{
		Encoding: intermediary.Encoding,
		Content:  data,
	}
	*b = newbody

	return nil
}

// New creates a new Mock from JSON bytes
func New(b []byte) (*Mock, error) {
	var mock Mock
	err := json.Unmarshal(b, &mock)

	//Assign ID
	mock.assignIdentifier()

	//Validate Logic / Set defaults
	mock.Instances = 1

	mock.environment = map[string]interface{}{
		"greet":    "Hello, %s!",
		"names":    []string{"world", "you"},
		"response": response{},
		"sprintf":  fmt.Sprintf,
	}

	code := `sprintf(greet, response.Body.Content)`
	mock.program, err = expr.Compile(code, expr.Env(mock.environment))
	if err != nil {
		panic(err)
	}

	return &mock, err
}

func (m *Mock) assignIdentifier() {
	var details []byte
	details = append(details, []byte(m.Name)...)
	details = append(details, []byte(m.Path)...)
	details = append(details, []byte(m.Selector)...)
	for _, rsp := range m.Responses {
		details = append(details, []byte(rsp.Body.Content)...)
		details = append(details, byte(rsp.Code.Value))
		for _, header := range rsp.Headers {
			details = append(details, []byte(header.Name)...)
			details = append(details, []byte(header.Value)...)
		}
	}
	hmd5 := md5.Sum(details)
	m.ID = fmt.Sprintf("%x", hmd5)
}

func (m *Mock) next() *response {
	if len(m.Responses) == 0 {
		// Set 404
		var rsp response = response{
			Body: body{Content: []byte("Not Found")},
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
	router.GET(m.Path, func(w http.ResponseWriter, r *http.Request, pm httprouter.Params) {
		start := time.Now()

		rsp := m.next()

		for _, h := range rsp.Headers {
			w.Header().Set(h.Name, h.Value)
		}
		w.WriteHeader(rsp.Code.Value)

		m.environment = map[string]interface{}{
			"greet":    "Bye, %s!",
			"names":    []string{"world", "you"},
			"response": rsp,
			"sprintf":  fmt.Sprintf,
		}

		output, err := expr.Run(m.program, m.environment)
		if err != nil {
			panic(err)
		}

		body := fmt.Sprint(output)

		desired := rsp.Delay.duration
		end := time.Now()
		elapsed := end.Sub(start)
		delay := desired - elapsed
		time.Sleep(delay)

		fmt.Fprintf(w, "%s", body)

	})
}
