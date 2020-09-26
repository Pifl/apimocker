package mock

import (
	"encoding/json"
	"errors"
	"strings"
)

type Mock struct {
	Id int
	Port int

	Name string
	Path string
	Selector selector

	Responses []Response
}

type Response struct {
	Body string
}

type selector string 
const (
	Sequence selector = "Sequence"
	Random = "Random"
	Script = "Script"
)

func (s *selector) UnmarshalJSON(b []byte) error {
    var tmp string
	json.Unmarshal(b, &tmp)
	tmp = strings.Title(strings.ToLower(tmp))
    selectorT := selector(tmp)
    switch selectorT {
		case Sequence, Random, Script:
			*s = selectorT
			return nil
		}
    return errors.New("Invalid Selector type")
}