package subrouter

import (
	"github.com/julienschmidt/httprouter"
	"log"
	"fmt"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"strconv"
	"apimocker/host"
	"apimocker/mock"
)

func AddHostsSubRouter(pathPrefix string, r *httprouter.Router) {
	path := "hosts"
	r.GET(pathPrefix + path + "/:port", getHostHandler)
	r.POST(pathPrefix + path, addMockHandler)
	r.DELETE(pathPrefix + path, removeMockHandler)
}

func getHostHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	port, err := strconv.Atoi(ps.ByName("port"))
	if err != nil {
		w.WriteHeader(403)
	}
	host := host.Host(port)

	rsp, err := json.Marshal(host)
	if err != nil {
		fmt.Println(err)
	}
	
	fmt.Fprintf(w, "%s\n", rsp)
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

	host := host.RegisterMock(mock)
	
	rsp, err := json.Marshal(host)
	if err != nil {
		fmt.Println(err)
	}
	
	fmt.Fprintf(w, "%s\n", rsp)

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

	host, err := host.RemoveMock(mock)
	if err != nil {
		fmt.Fprintln(w, err)
	}

	rsp, err := json.Marshal(host)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Fprintf(w, "%s\n", rsp)

}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}