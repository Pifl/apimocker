package routers


import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"apimocker/host"
	"fmt"
)


func AddUtilSubRouter(pathPrefix string, r *httprouter.Router) {
	path := "util"
	r.GET(pathPrefix + path, utilHandler)
}

func utilHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	host.Counter = host.Counter + 1
    fmt.Fprintf(w, "%v", host.Counter)
}
