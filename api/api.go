package api

import (
	"log"
	"net/http"
	"github.com/julienschmidt/httprouter"
	"apimocker/api/subrouter"
)

func Start(port string){
	r := httprouter.New()

	pathPrefix := "/api/v1/"

	subrouter.AddHostsSubRouter(pathPrefix, r)

	log.Fatal(http.ListenAndServe(port, ContentType(r)))
}

func ContentType(handler http.Handler) http.Handler {
    setContent := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
        handler.ServeHTTP(w, r)
    }
    return http.HandlerFunc(setContent)
}
