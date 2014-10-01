// Copyright (c) 2014 Datacratic. All rights reserved.

package rest

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// Endpoint routes incoming bid requests to the registered routes. Implements
// the http.Handler interface and can also act as it's own standalone server.
//
// The endpoint currently only supports JSON content-type for regular message
// and text/plain for error messages.
type Endpoint struct {
	*http.Server

	// Root is the path prefix of all the routes to be matched by this
	// endpoint. Must be set before calling Init and can't be changed
	// afterwards.
	Root string

	// ErrorFunc is called for all errors that passes through this endpoint. The
	// returned value will overwrite the current error and will be returned to
	// the client instead.. If it's return value is a rest.CodedError then the
	// status code of the error object will be used.
	ErrorFunc func(ErrorType, error) error

	initialize sync.Once

	router router
}

// Init initializes the object.
func (endpoint *Endpoint) Init() {
	endpoint.initialize.Do(endpoint.init)
}

func (endpoint *Endpoint) init() {
	if len(endpoint.Root) == 0 {
		endpoint.Root = "/"
	} else {
		endpoint.Root = "/" + strings.Trim(endpoint.Root, "/")
	}

	if endpoint.Server == nil {
		endpoint.Server = &http.Server{}
	}
}

// AddRoute adds all the given routes to the endpoint.
func (endpoint *Endpoint) AddRoute(routes ...*Route) {
	endpoint.Init()

	for _, route := range routes {
		endpoint.router.Add(route)
	}
}

// AddRoutable adds all the routes returned by the Routable objects to the
// endpoint.
func (endpoint *Endpoint) AddRoutable(routables ...Routable) {
	for _, routable := range routables {
		endpoint.AddRoute(routable.RESTRoutes()...)
	}
}

func (endpoint *Endpoint) initServer() {
	endpoint.Init()

	if endpoint.Handler != nil {
		panic("Handler must be nil")
	}

	endpoint.Handler = endpoint
}

// Serve is a non-blocking wrapper for http.Server.Serve which panics if an
// error occurs.
func (endpoint *Endpoint) Serve(listener net.Listener) {
	endpoint.initServer()
	go func() {
		if err := endpoint.Server.Serve(listener); err != nil {
			panic(err.Error())
		}
	}()
}

// ListenAndServe is a non-blocking wrapper for http.Server.ListenAndServe which
// panics if an error occurs.
func (endpoint *Endpoint) ListenAndServe() {
	endpoint.initServer()
	go func() {
		if err := endpoint.Server.ListenAndServe(); err != nil {
			panic(err.Error())
		}
	}()
}

// ListenAndServeTLS is a non-blocking wrapper for http.Server.ListenAndServeTLS
// which panics if an error occurs.
func (endpoint *Endpoint) ListenAndServeTLS(certFile, keyFile string) {
	endpoint.initServer()
	go func() {
		if err := endpoint.Server.ListenAndServeTLS(certFile, keyFile); err != nil {
			panic(err.Error())
		}
	}()
}

func (endpoint *Endpoint) route(method, path string) (*Route, []string, error) {
	if strings.HasPrefix(path, endpoint.Root) {
		sub := path[len(endpoint.Root):]
		if route, args := endpoint.router.Route(method, sub); route != nil {
			return route, args, nil
		}
	}

	return nil, nil, fmt.Errorf("unknown path: '%s'", path)
}

func (endpoint *Endpoint) respondError(writer http.ResponseWriter, errType ErrorType, code int, err error) {
	if endpoint.ErrorFunc != nil {
		err = endpoint.ErrorFunc(errType, err)
	}

	if coded, ok := err.(*CodedError); ok {
		code = coded.Code
		err = coded.Sub
	}

	http.Error(writer, err.Error(), code)
}

// ServeHTTP services incoming HTTP request by routing them to one of the
// registered routes. Handles all marshalling of input and outputs as well as
// any required path parsing.
func (endpoint *Endpoint) ServeHTTP(writer http.ResponseWriter, httpReq *http.Request) {
	endpoint.Init()

	route, args, err := endpoint.route(httpReq.Method, httpReq.URL.Path)
	if err != nil {
		endpoint.respondError(writer, UnknownRoute, http.StatusNotFound, err)
		return
	}

	if contentType := httpReq.Header.Get("Content-Type"); contentType != "application/json" {
		err := fmt.Errorf("unsupported content type: got '%s' expected 'application/json'", contentType)
		endpoint.respondError(writer, UnsupportedContentType, http.StatusBadRequest, err)
		return
	}

	body, err := ioutil.ReadAll(httpReq.Body)
	if err != nil {
		endpoint.respondError(writer, ReadBodyError, http.StatusBadRequest, err)
		return
	}

	resp, restError := route.invoke(args, body)
	if restError != nil {
		endpoint.respondError(writer, restError.Type, http.StatusBadRequest, restError.Sub)
		return
	}

	if len(resp) == 0 {
		writer.WriteHeader(http.StatusNoContent)
	} else {
		header := writer.Header()
		header.Set("Content-Type", "application/json")
		header.Set("Content-Length", strconv.FormatInt(int64(len(resp)), 10))
		writer.Write(resp)
	}
}

// TestEndpoint is a convenience endpoint used to create non-conflicting
// temporary endpoints in tests.
type TestEndpoint struct {
	Endpoint
	Listener net.Listener
}

// Addr returns the address of the endpoint.
func (endpoint *TestEndpoint) Addr() net.Addr {
	return endpoint.Listener.Addr()
}

// URL returns the URL where the endpoint can be reached. Does not include the
// Root of the endpoint in the URL.
func (endpoint *TestEndpoint) URL() string {
	return "http://" + endpoint.Addr().String()
}

// RootedURL returns the URL, including the root,  where the endpoint can be reached.
func (endpoint *TestEndpoint) RootedURL() string {
	return "http://" + endpoint.Addr().String() + endpoint.Root
}

// ListenAndServe is a non-blocking wrapper function for Endpoint.ListenAndServe
// which allocates a random free port on localhost for the endpoint.
func (endpoint *TestEndpoint) ListenAndServe() (err error) {
	endpoint.Listener, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		endpoint.Listener, err = net.Listen("tcp6", "[::1]:0")
	}

	if err == nil {
		endpoint.Endpoint.Serve(endpoint.Listener)
	}

	return
}

// Serve is not supported on this endpoint and will panic if called.
func (endpoint *TestEndpoint) Serve(listener net.Listener) error {
	panic("unsupported")
}

// ListenAndServeTLS is not supported on this endpoint and will panic if called.
func (endpoint *TestEndpoint) ListenAndServeTLS(certFile, keyFile string) error {
	panic("unsupported")
}
