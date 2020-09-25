package api

import (
	"log"
	"net/http"
	"github.com/julienschmidt/httprouter"
	"apimocker/api/routers"
)

func Start(port string){
	router := httprouter.New()
	pathPrefix := "/api/v1/"
	routers.AddAdminSubRouter(pathPrefix, router)
	routers.AddUtilSubRouter(pathPrefix, router)
	routers.AddMockSubRouter(pathPrefix, router)

	log.Fatal(http.ListenAndServe(port, router))
}