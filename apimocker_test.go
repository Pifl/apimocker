package main

import (
	"apimocker/host"
	"apimocker/mock"
	"encoding/json"
	"testing"
)

func TestCreateHostAndMock(t *testing.T) {

	var mock mock.Mock
	var request = `{
		"Name": "Typical Scenario",
		"path": "/example",
		"selector": "sequence",
		"responses": [
			{
				"body": "Hello, World!"
			},
			{
				"body": "Goodbye, World!"
			}
		]
	}`
	err := json.Unmarshal([]byte(request), &mock)
	if err != nil {
		t.Error("JSON Unmarshalling failed, test setup is wrong")
		return
	}

	host, _ := host.RegisterMock(5500, &mock)
	if len(host.Mocks) != 1 {
		t.Errorf("Tried to add a single mock but length is %v", len(host.Mocks))
	}
}
