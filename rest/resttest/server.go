// Copyright (c) 2014 Datacratic. All rights reserved.

package resttest

import (
	"github.com/datacratic/gorest/rest"

	"net/http/httptest"
)

// NewServer creates an httptest.Server that uses the given route as the REST
// handler.
func NewServer(route *rest.Route) *httptest.Server {
	endpoint := new(rest.Mux)
	endpoint.AddRoute(route)

	return httptest.NewServer(endpoint)
}

// NewService creates an httptest.Server that uses the given service as the REST
// handler.
func NewService(service rest.Routable) *httptest.Server {
	endpoint := new(rest.Mux)
	endpoint.AddService(service)

	return httptest.NewServer(endpoint)
}
