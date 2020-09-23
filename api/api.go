package api

import (
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"apimocker/api/routers"
)

func Start(port string){
	router := mux.NewRouter()
	routers.AddAdminSubRouter(router)
	routers.AddUtilSubRouter(router)

	log.Fatal(http.ListenAndServe(port, router))
}