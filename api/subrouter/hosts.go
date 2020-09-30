package subrouter

import (
	"apimocker/host"
	"apimocker/mock"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// AddHostsSubRouter adds handlers to the sub paths for "hosts"
func AddHostsSubRouter(pathPrefix string, r *httprouter.Router) {
	path := "hosts"
	r.GET(pathPrefix+path+"/:port", getHostHandler)
	r.GET(pathPrefix+path+"/:port/mock/:id", getMockHandler)
	r.POST(pathPrefix+path+"/:port", addMockHandler)
	r.DELETE(pathPrefix+path+"/:port/mock/:id", removeMockHandler)
}

func getHostHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	port, err := strconv.Atoi(ps.ByName("port"))
	if err != nil {
		w.WriteHeader(400)
	}

	host, ok := host.ByPort(port)
	if !ok {
		w.WriteHeader(400)
	}

	rsp, err := json.Marshal(host)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Fprintf(w, "%s\n", rsp)
}

func getMockHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	port, err := strconv.Atoi(ps.ByName("port"))
	if err != nil {
		w.WriteHeader(400)
	}

	id := ps.ByName("id")

	host, ok := host.ByPort(port)
	if !ok {
		w.WriteHeader(400)
	}

	mock, err := host.Mock(id)
	if err != nil {
		w.WriteHeader(400)
	}

	rsp, err := json.Marshal(mock)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Fprintf(w, "%s\n", rsp)

}
func addMockHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	port, err := strconv.Atoi(ps.ByName("port"))
	if err != nil {
		w.WriteHeader(400)
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	mock, err := mock.New(body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	mock.Port = port

	fmt.Println(mock.Responses[0].Body)

	host := host.RegisterMock(mock)

	rsp, err := json.Marshal(host)
	if err != nil {
		fmt.Println(err)
	}

	w.WriteHeader(201)
	fmt.Fprintf(w, "%s\n", rsp)

}

func removeMockHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	port, err := strconv.Atoi(ps.ByName("port"))
	if err != nil {
		w.WriteHeader(400)
	}

	id := ps.ByName("id")

	host, err := host.RemoveMock(port, id)
	if err != nil {
		fmt.Fprintln(w, err)
	}

	rsp, err := json.Marshal(host)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Fprintf(w, "%s\n", rsp)

}
