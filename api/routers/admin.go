package routers

import (
	"github.com/gorilla/mux"
	"net/http"
)


func AddAdminSubRouter(r *mux.Router) {
	s := r.PathPrefix("/admin").Subrouter()
	s.HandleFunc("", adminHandler)
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Gorilla!\n"))
}
