package api

import (
	"github.com/gorilla/mux"
)

func setupRoutes(r *mux.Router){
	r.HandleFunc("/books/{title}/page/{page}", adminHandler())
}