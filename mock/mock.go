package mock

import (
	"crypto/md5"
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
	ID string

	Name     string
	Path     string
	Selector selector

	index     int
	Responses []*response

	Instances int
}

type response struct {
	Body    body
	Code    code
	Headers []header
	Delay   delay
}

type body struct {
	encoding encoding
	content  string
}
type encoding string

const (
	base64 encoding = "base64"
	base32          = "base32"
	raw             = "raw"
)

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

func (b *body) UnmarshalJSON(bytes []byte) error {
	tokens := make([]string, 0, 4)
	token := false
	var start int
	for i, b := range bytes {
		// If current byte equals a qotation mark signifies start of token
		if b == 34 {
			if !token {
				token = true
				start = i + 1
			} else {
				token = false
				tokens = append(tokens, string(bytes[start:i]))
			}
		}
	}
	//If the first token is encoding then the second must be the value, else last must be the value
	var encod string
	var content string
	if strings.EqualFold(tokens[0], "encoding") {
		encod = tokens[1]
		content = tokens[3]
	} else {
		encod = tokens[3]
		content = tokens[1]
	}

	// Check if encoding is one of the accepted values
	encodT := encoding(encod)
	switch encodT {
	case base64, base32:
	default:
		encodT = raw
	}

	newbody := body{
		encoding: encodT,
		content:  content,
	}
	fmt.Printf("%v", newbody)
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
	return &mock, err
}

func (m *Mock) assignIdentifier() {
	var details []byte
	details = append(details, []byte(m.Name)...)
	details = append(details, []byte(m.Path)...)
	details = append(details, []byte(m.Selector)...)
	for _, rsp := range m.Responses {
		details = append(details, []byte(rsp.Body.content)...)
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
			Body: body{content: "Not Found"},
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

		fmt.Fprint(w, rsp.Body.content+"\n")

	})
}
