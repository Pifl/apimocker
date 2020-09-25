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

var hosts map[int]*host = make(map[int]*host)

type host struct {
	Port int
	Mocks []mock.Mock
	Server *http.Server `json:"-"`
}

type SyntaxError struct {
    msg    string // description of error
    Offset int64  // error occurred after reading Offset bytes
}

func (e *SyntaxError) Error() string { 
	return e.msg 
}


func (host *host) AddMock(mock mock.Mock) {
	host.Mocks = append(host.Mocks, mock)
}

func New(port int) *host {
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

	return &host {
		Port: port,
		Mocks: mocks,
		Server: srv,
	}
}

func Host(port int) *host {
	host, ok := hosts[port]
	if !ok {
		return nil
	}
	return host
}

func Mock(port, id int) (*mock.Mock, error){
	host, ok := hosts[port]
	if !ok {
		return nil, &SyntaxError{}
	}

	for i, m := range host.Mocks {
		if m.Id == id {
			return &host.Mocks[i], nil
		}
	}

	return nil, &SyntaxError{}
}

func RemoveMock(port, id int) (*host, error) {
	host, ok := hosts[port]
	if !ok {
		return nil, &SyntaxError{}
	}
	
	for i, m := range host.Mocks {
		if m.Id == id {
			host.Mocks[i] = host.Mocks[len(host.Mocks) - 1]
			host.Mocks = host.Mocks[:len(host.Mocks) - 1]
			break
		}
	}

	//If no mocks left, shutdown host
	if len(host.Mocks) == 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    	defer cancel()
		if err := host.Server.Shutdown(ctx); err != nil {
			panic(err) 
		}
		delete(hosts, port)
	}
	return host, nil
}

func RegisterMock(mock mock.Mock) *host {
	host, ok := hosts[mock.Port]
	if !ok {
		host = New(mock.Port)
		hosts[mock.Port] = host
	}
	assignIdentifier(host, &mock)
	host.AddMock(mock)
	return host
}

func assignIdentifier(h *host, m *mock.Mock) {
	id := m.Port << 8 + int(len(h.Mocks))
	m.Id = id
}