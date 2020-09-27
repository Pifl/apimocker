package api

import (
	"apimocker/api/subrouter"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Start runs the API Server with port
func Start(port string) {
	r := httprouter.New()

	pathPrefix := "/api/v1/"

	subrouter.AddHostsSubRouter(pathPrefix, r)

	log.Fatal(http.ListenAndServe(port, contentType(r)))
}

func contentType(handler http.Handler) http.Handler {
	setContent := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		handler.ServeHTTP(w, r)
	}
	return http.HandlerFunc(setContent)
}
