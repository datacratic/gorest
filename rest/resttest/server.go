// Copyright (c) 2014 Datacratic. All rights reserved.

package resttest

import (
	"github.com/datacratic/gorest/rest"

	"net/http/httptest"
)

func NewServer(route *rest.Route) *httptest.Server {
	endpoint := new(rest.Mux)
	endpoint.AddRoute(route)

	return httptest.NewServer(endpoint)
}

func NewService(routable rest.Routable) *httptest.Server {
	endpoint := new(rest.Mux)
	endpoint.AddService(routable)

	return httptest.NewServer(endpoint)
}
