package mock

import (
	"crypto/md5"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
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
	Path     path
	Selector selector

	index       int
	environment map[string]interface{}
	program     *vm.Program
	Responses   []*response

	Instances int
}

type path struct {
	Method   string
	Resource string
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
)

type Env struct {
	Request request
}

func (p *path) UnmarshalJSON(b []byte) error {
	var tmp string
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}
	var segs = strings.Split(tmp, " ")
	if len(segs) != 2 {
		return errors.New("Invalid Path field, expected format METHOD /resource, e.g. GET /example")
	}
	method := strings.ToUpper(segs[0])
	resource := segs[1]

	switch method {
	case "GET", "POST":
	default:
		return fmt.Errorf("Unsupported METHOD type %v", method)
	}

	path := path{
		Method:   method,
		Resource: resource,
	}
	*p = path
	return nil
}

func (p path) MarshalJSON() ([]byte, error) {
	path := p.Method + " " + p.Resource
	b, err := json.Marshal(path)
	return b, err
}

func (s *selector) UnmarshalJSON(b []byte) error {
	var tmp string
	json.Unmarshal(b, &tmp)
	selectorT := selector(strings.Title(strings.ToLower(tmp)))
	switch selectorT {
	case sequence, random:
		*s = selectorT
		return nil
	default:
		fmt.Println(tmp)

		_, err := expr.Compile(tmp, expr.Env(Env{}))
		if err != nil {
			fmt.Println(err)
			return errors.New("Invalid Selector script")
		}
		*s = selector(tmp)
		return nil
	}
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
	var m Mock
	err := json.Unmarshal(b, &m)

	//Assign ID
	m.assignIdentifier()

	//Validate Logic / Set defaults
	m.Instances = 1

	if m.Selector != random && m.Selector != sequence {

		m.program, err = expr.Compile(string(m.Selector), expr.Env(Env{}))
		if err != nil {
			fmt.Println(err)
			return nil, errors.New("Invalid Selector script")
		}
	}

	return &m, err
}

func (m *Mock) assignIdentifier() {
	var details []byte
	details = append(details, []byte(m.Name)...)
	details = append(details, []byte(m.Path.Method)...)
	details = append(details, []byte(m.Path.Resource)...)
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

type request struct {
	Body   string
	Params httprouter.Params
}

func (m *Mock) next(req request) *response {
	if len(m.Responses) == 0 {
		// Set 404
		var rsp response = response{
			Code: code{Value: 404},
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
		env := Env{
			Request: req,
		}
		output, err := expr.Run(m.program, env)
		if err != nil {
			panic(err)
		}
		// Format output as string
		soutput := fmt.Sprintf("%v", output)
		// Assert output type as int
		selection, err := strconv.Atoi(soutput)
		if err != nil || selection >= len(m.Responses) {
			var rsp response = response{
				Code: code{Value: 500},
				Body: body{Content: []byte(fmt.Sprintf("Incorrect output from selection expression: %v \n %v", selection, err))},
			}
			return &rsp
		}
		i = selection
	}
	m.index = i
	return m.Responses[m.index]
}

// Handler adds a mock specifc handler to the router of the host server
func (m *Mock) Handler(router *httprouter.Router) {
	router.Handle(m.Path.Method, m.Path.Resource, func(w http.ResponseWriter, r *http.Request, pm httprouter.Params) {
		start := time.Now()

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}

		req := request{
			Body:   string(body),
			Params: pm,
		}

		rsp := m.next(req)

		for _, h := range rsp.Headers {
			w.Header().Set(h.Name, h.Value)
		}
		w.WriteHeader(rsp.Code.Value)

		desired := rsp.Delay.duration
		end := time.Now()
		elapsed := end.Sub(start)
		delay := desired - elapsed
		time.Sleep(delay)

		fmt.Fprintf(w, "%s", rsp.Body.Content)

	})
}
