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
	path := "host"
	r.GET(pathPrefix+path+"/:port", getHostHandler)
	r.GET(pathPrefix+path+"/:port/mock/:id", getMockHandler)
	r.POST(pathPrefix+path+"/:port", addMockHandler)
	r.DELETE(pathPrefix+path+"/:port/mock/:id", removeMockHandler)
}

// ../host/{port}
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

// ../host/{port}/mock/{mock.id}
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

	mock, _ := host.MockByID(id)
	if mock == nil {
		w.WriteHeader(400)
	}

	rsp, err := json.Marshal(mock)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Fprintf(w, "%s\n", rsp)

}

// ../host/{port}
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

	_, err = host.RegisterMock(port, mock)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusBadRequest)
		return
	}

	rsp, err := json.Marshal(mock)
	if err != nil {
		fmt.Println(err)
	}

	w.WriteHeader(201)
	fmt.Fprintf(w, "%s\n", rsp)

}

// ../host/{port}/mock/{mock.id}?force={force}
// If force then remove the mock from the host
// else decrement instances and only remove mock if instances
// is equal to zero
func removeMockHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	port, err := strconv.Atoi(ps.ByName("port"))
	if err != nil {
		w.WriteHeader(400)
	}
	id := ps.ByName("id")

	force := (r.URL.Query().Get("force") == "true")

	host, err := host.RemoveMock(port, id, force)
	if err != nil {
		fmt.Fprintln(w, err)
	}

	rsp, err := json.Marshal(host)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Fprintf(w, "%s\n", rsp)

}
