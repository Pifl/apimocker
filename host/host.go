package host

import (
	"apimocker/mock"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

var hosts map[int]*Host = make(map[int]*Host)

// Host holds a Server and the Mocks together
type Host struct {
	Port   int
	Router *httprouter.Router `json:"-"`
	Mocks  []mock.Mock
	Server *http.Server `json:"-"`
}

type syntaxError struct {
	msg    string // description of error
	Offset int64  // error occurred after reading Offset bytes
}

func (e *syntaxError) Error() string {
	return e.msg
}

func (host *Host) addMock(mock mock.Mock) {
	//GET HANDLER FROM MOCK ADD TO HOST ROUTER
	/*
		router.GET("/", func (w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			fmt.Fprint(w, "Welcome!\n")
		})
	*/
	mock.Handler(host.Router)
	host.Mocks = append(host.Mocks, mock)
}

// New initalises a host including starting a server
func New(port int) *Host {
	mocks := make([]mock.Mock, 0, 5)
	router := httprouter.New()

	var srv *http.Server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	return &Host{
		Port:   port,
		Mocks:  mocks,
		Server: srv,
		Router: router,
	}
}

// ByPort gets the host on a port
func ByPort(port int) *Host {
	host, ok := hosts[port]
	if !ok {
		return nil
	}
	return host
}

// Mock gets a mock on a port with the id
func Mock(port, id int) (*mock.Mock, error) {
	host, ok := hosts[port]
	if !ok {
		return nil, &syntaxError{}
	}

	for i, m := range host.Mocks {
		if m.ID == id {
			return &host.Mocks[i], nil
		}
	}

	return nil, &syntaxError{}
}

// RemoveMock from a host using the id
func RemoveMock(port, id int) (*Host, error) {
	host, ok := hosts[port]
	if !ok {
		return nil, &syntaxError{}
	}

	for i, m := range host.Mocks {
		if m.ID == id {
			host.Mocks[i] = host.Mocks[len(host.Mocks)-1]
			host.Mocks = host.Mocks[:len(host.Mocks)-1]
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

// RegisterMock to a host, creates a new host if one not available
func RegisterMock(mock mock.Mock) *Host {
	host, ok := hosts[mock.Port]
	if !ok {
		host = New(mock.Port)
		hosts[mock.Port] = host
	}
	assignIdentifier(host, &mock)
	host.addMock(mock)
	return host
}

func assignIdentifier(h *Host, m *mock.Mock) {
	id := m.Port<<8 + int(len(h.Mocks))
	m.ID = id
}
