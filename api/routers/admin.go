package routers

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)


func AddAdminSubRouter(pathPrefix string, r *httprouter.Router) {
	path := "admin"
	r.GET(pathPrefix + path, adminHandler)
}

func adminHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    w.Write([]byte("Gorilla!\n"))
}
