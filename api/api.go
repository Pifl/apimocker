package api

import (
	"github.com/gorilla/mux"
	"net/http"
)

func Start(port string) {
	r := mux.NewRouter()

	setupRoutes(r);

	http.ListenAndServe(":5050", r)
}

