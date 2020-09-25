package routers

import (
	"github.com/julienschmidt/httprouter"
	"log"
	"fmt"
	"encoding/json"
	"net/http"
	"io/ioutil"

	"apimocker/host"
	"apimocker/mock"
)

func AddMockSubRouter(pathPrefix string, r *httprouter.Router) {
	path := "mocks"
	r.POST(pathPrefix + path, addMockHandler)
	r.DELETE(pathPrefix + path, removeMockHandler)
}

func addMockHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	var mock mock.Mock
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	json.Unmarshal(body, &mock)

	host.RegisterMock(mock)

    // router := httprouter.New()
	// router.GET("/", MockHandler(request.Text))
	
	// go http.ListenAndServe(fmt.Sprintf(":%d", request.Port), router)
	// fmt.Fprintf(w, "Mock Started on Port: %d\n", request.Port);
}

func removeMockHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var mock mock.Mock
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	json.Unmarshal(body, &mock)

	host.RemoveMock(mock)
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}