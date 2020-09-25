package host

import (
	"fmt"
	"log"
	"time"
	"context"
	"apimocker/mock"
	"net/http"

	"github.com/julienschmidt/httprouter"
)
var Counter int = 0
var hosts map[int]*Host = make(map[int]*Host)

type Host struct {
	Mocks []mock.Mock
	Server *http.Server
}

type SyntaxError struct {
    msg    string // description of error
    Offset int64  // error occurred after reading Offset bytes
}

func (e *SyntaxError) Error() string { 
	return e.msg 
}


func (host *Host) AddMock(mock mock.Mock) {
	host.Mocks = append(host.Mocks, mock)
}

func New(port int) *Host {
	mocks := make([]mock.Mock,0,5)
	router := httprouter.New()

	router.GET("/", func (w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		fmt.Fprint(w, "Welcome!\n")
	})
	
	var srv *http.Server = &http.Server{
		Addr: fmt.Sprintf(":%d", port),
		Handler: router,
	}

	
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatalf("ListenAndServe(): %v", err)
        }
	}()

	return &Host {
		Mocks: mocks,
		Server: srv,
	}
}


func RemoveMock(mock mock.Mock) error {
	host, ok := hosts[mock.Port]
	if !ok {
		return &SyntaxError{}
	}
	//Just remove the first one for now will reference stuff properly later
	host.Mocks[0] = host.Mocks[len(host.Mocks) - 1]
	host.Mocks = host.Mocks[:len(host.Mocks) - 1]

	if len(host.Mocks) == 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    	defer cancel()
		if err := host.Server.Shutdown(ctx); err != nil {
			panic(err) 
		}
		delete(hosts, mock.Port)
	}
	return nil
}

func RegisterMock(mock mock.Mock) {
	host, ok := hosts[mock.Port]
	if !ok {
		host = New(mock.Port)
		hosts[mock.Port] = host
	}
	host.AddMock(mock)
}