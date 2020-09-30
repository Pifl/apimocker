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
	Mocks  []*mock.Mock
	Server *http.Server `json:"-"`
}

type hostError struct {
	msg string
}

func (err *hostError) Error() string {
	return err.msg
}

type mergeError struct {
	msg string
}

func (err *mergeError) Error() string {
	return err.msg
}

// Add new mock to an existing host
// For desired functionality that if you request two of the same mock
// to delete it you need to delete it twice (or force delete)
// We check if mock with the same path already existings, if it does then we check
// if the mock is identical / if so increment Instances, if not then try to "merge"
func (host *Host) addMock(mock *mock.Mock) error {
	existingMock, _ := host.MockByPath(mock.Path)
	if existingMock == nil {
		mock.Handler(host.Router)
		host.Mocks = append(host.Mocks, mock)
		return nil
	}
	// ID is a hash of the contents of the mock, if ID is the same assume contents is identical
	if existingMock.ID == mock.ID {
		existingMock.Instances++
		mock.Instances = existingMock.Instances
		return nil
	}
	// Want to use the same path but aren't identical
	// TODO Handle merging
	return &mergeError{msg: "The new Mock wants to occupy the same resource but can't merge"}
}

// New initalises a host including starting a server
func New(port int) *Host {
	mocks := make([]*mock.Mock, 0, 5)
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

// MockByPath get a mock on a host with a path
func (host *Host) MockByPath(path string) (*mock.Mock, int) {
	for i, m := range host.Mocks {
		if m.Path == path {
			return host.Mocks[i], i
		}
	}
	return nil, -1
}

// MockByID gets a mock on a host with the id
func (host *Host) MockByID(id string) (*mock.Mock, int) {
	for i, m := range host.Mocks {
		if m.ID == id {
			return host.Mocks[i], i
		}
	}
	return nil, -1
}

// RemoveMock from a host using the id
func RemoveMock(port int, id string, byforce bool) (*Host, error) {
	host, ok := ByPort(port)
	if !ok {
		return nil, &hostError{msg: "A host does not exist on this port"}
	}

	m, i := host.MockByID(id)
	if m == nil {
		return nil, &hostError{msg: "A mock does not exist with this id"}
	}

	//Check if the instances will go down to zero, if so remove mock
	instances := m.Instances - 1
	if instances == 0 || byforce {
		host.Mocks[i] = host.Mocks[len(host.Mocks)-1]
		host.Mocks = host.Mocks[:len(host.Mocks)-1]
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
func RegisterMock(port int, mock *mock.Mock) (*Host, error) {
	host, ok := hosts[port]
	if !ok {
		host = New(port)
		hosts[port] = host
	}
	err := host.addMock(mock)
	return host, err
}
