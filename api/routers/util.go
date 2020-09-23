package routers


import (
	"github.com/gorilla/mux"
	"net/http"
	"apimocker/host"
	"fmt"
)


func AddUtilSubRouter(r *mux.Router) {
	s := r.PathPrefix("/util").Subrouter()
	s.HandleFunc("", utilHandler)
}

func utilHandler(w http.ResponseWriter, r *http.Request) {
	host.Counter = host.Counter + 1
    fmt.Fprintf(w, "%v", host.Counter)
}
