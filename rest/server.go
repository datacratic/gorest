// Copyright (c) 2014 Datacratic. All rights reserved.

package rest

import (
	"net"
	"net/http"
)

// DefaultMux is the default Mux used by Serve which uses the
// http.DefaultServeMux as the DefaultHandler in Mux if no routes match.
var DefaultMux = new(Mux)

// AddRoute adds a REST route to DefaultMux. See Route type for further details
// on REST route specification.
func AddRoute(path, method string, handler interface{}) {
	DefaultMux.AddRoute(NewRoute(path, method, handler))
}

// AddService adds a Routable service to DefaultMux.
func AddService(routable Routable) {
	DefaultMux.AddService(routable)
}

// Serve is a proxy for the http.Serve function but using the DefaultMux.
func Serve(l net.Listener, mux *Mux) error {
	if mux == nil {
		mux = DefaultMux
	}

	srv := &http.Server{Handler: mux}
	return srv.Serve(l)
}

// ListenAndServe is a proxy for the http.ListenAndServe function but using the
// DefaultMux.
func ListenAndServe(addr string, mux *Mux) error {
	if mux == nil {
		mux = DefaultMux
	}

	server := &http.Server{Addr: addr, Handler: mux}
	return server.ListenAndServe()
}

// ListenAndServeTLS is a proxy for the http.ListenAndServeTLS function but
// using the DefaultMux.
func ListenAndServeTLS(addr string, certFile string, keyFile string, mux *Mux) error {
	if mux == nil {
		mux = DefaultMux
	}

	server := &http.Server{Addr: addr, Handler: mux}
	return server.ListenAndServeTLS(certFile, keyFile)
}
