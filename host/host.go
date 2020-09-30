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
	Mocks  map[string]*mock.Mock
	Server *http.Server `json:"-"`
}

type syntaxError struct {
	msg    string // description of error
	Offset int64  // error occurred after reading Offset bytes
}

func (e *syntaxError) Error() string {
	return e.msg
}

// Add new mock to an existing host
// For desired functionality that if you request two of the same mock
// to delete it you need to delete it twice (or force delete)
// We check if mock with the same path already existings, if it does then we check
// if the mock is identical / if so increment Instances, if not then try to "merge"
func (host *Host) addMock(mock *mock.Mock) {
	existingMock, ok := host.Mocks[mock.Path]
	if !ok {
		mock.Handler(host.Router)
		host.Mocks[mock.Path] = mock

		return
	}
	// TODO Handle merging
	existingMock.Instances++
}

// New initalises a host including starting a server
func New(port int) *Host {
	mocks := make(map[string]*mock.Mock)
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
func ByPort(port int) (*Host, bool) {
	host, ok := hosts[port]
	return host, ok
}

// Mock gets a mock on a host with the id
func (host *Host) Mock(id string) (*mock.Mock, error) {
	for i, m := range host.Mocks {
		if m.ID == id {
			return host.Mocks[i], nil
		}
	}
	return nil, &syntaxError{}
}

// RemoveMock from a host using the id
func RemoveMock(port int, id string) (*Host, error) {
	host, ok := ByPort(port)
	if !ok {
		return nil, &syntaxError{}
	}

	m, err := host.Mock(id)
	if err != nil {
		return nil, &syntaxError{}
	}

	//Check if the instances will go down to zero, if so remove mock
	instances := m.Instances - 1
	if instances == 0 {
		delete(host.Mocks, m.Path)
	} else {
		m.Instances = instances
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
func RegisterMock(mock *mock.Mock) *Host {
	host, ok := hosts[mock.Port]
	if !ok {
		host = New(mock.Port)
		hosts[mock.Port] = host
	}
	host.addMock(mock)
	return host
}
