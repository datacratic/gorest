// Copyright (c) 2014 Datacratic. All rights reserved.

package resttest

import (
	"github.com/datacratic/gorest/rest"

	"net/http/httptest"
)

// Server is a light utility wrapper around httptest.Server which provides a
// useful RootedURL function.
type Server struct {
	*httptest.Server
	mux *rest.Mux
}

// RootedURL returns the URL of the server plus the Root of the underlying
// rest.Mux.
func (server *Server) RootedURL() string {
	return rest.JoinPath(server.URL, server.mux.Root)
}

func newServer(mux *rest.Mux) *Server {
	return &Server{httptest.NewServer(mux), mux}
}

// NewServer creates an httptest.Server that uses the given route as the REST
// handler.
func NewServer(routes ...*rest.Route) *Server {
	return NewRootedServer("", routes...)
}

// NewRootedServer creates an httptest.Server that uses the given route as the
// REST handler rooted at the given path.
func NewRootedServer(root string, routes ...*rest.Route) *Server {
	mux := &rest.Mux{Root: root}
	for _, route := range routes {
		mux.AddRoute(route)
	}
	return newServer(mux)
}

// NewService creates an httptest.Server that uses the given service as the REST
// handler.
func NewService(service ...rest.Routable) *Server {
	return NewRootedService("", service...)
}

// NewRootedService creates an httptest.Server that uses the given service as
// the REST handler rooted at the given path.
func NewRootedService(root string, service ...rest.Routable) *Server {
	mux := &rest.Mux{Root: root}
	mux.AddService(service...)
	return newServer(mux)
}
