// Copyright (c) 2014 Datacratic. All rights reserved.

package rest

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// Endpoint routes incoming bid requests to the registered routes. Implements
// the http.Handler interface.
//
// The endpoint currently only supports JSON content-type for regular message
// and text/plain for error messages.
type Endpoint struct {

	// Root is the path prefix of all the routes to be matched by this
	// endpoint. Must be set before calling Init and can't be changed
	// afterwards.
	Root string

	// ErrorFunc is called for all errors that passes through this endpoint. The
	// returned value will overwrite the current error and will be returned to
	// the client instead.. If it's return value is a rest.CodedError then the
	// status code of the error object will be used.
	ErrorFunc func(ErrorType, error) error

	DefaultHandler http.Handler

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

	if endpoint.DefaultHandler == nil {
		endpoint.DefaultHandler = http.DefaultServeMux
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
func (endpoint *Endpoint) AddService(routables ...Routable) {
	for _, routable := range routables {
		endpoint.AddRoute(routable.RESTRoutes()...)
	}
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
		log.Printf("using default handler for '%s'", httpReq.URL.Path)
		endpoint.DefaultHandler.ServeHTTP(writer, httpReq)
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

var DefaultEndpoint = new(Endpoint)

func AddRoute(method, path string, handler interface{}) {
	DefaultEndpoint.AddRoute(NewRoute(method, path, handler))
}

func AddService(routable Routable) {
	DefaultEndpoint.AddService(routable)
}

func Serve(l net.Listener, endpoint *Endpoint) error {
	if endpoint == nil {
		endpoint = DefaultEndpoint
	}

	srv := &http.Server{Handler: endpoint}
	return srv.Serve(l)
}

func ListenAndServe(addr string, endpoint *Endpoint) error {
	if endpoint == nil {
		endpoint = DefaultEndpoint
	}

	server := &http.Server{Addr: addr, Handler: endpoint}
	return server.ListenAndServe()
}

func ListenAndServeTLS(addr string, certFile string, keyFile string, endpoint *Endpoint) error {
	if endpoint == nil {
		endpoint = DefaultEndpoint
	}

	server := &http.Server{Addr: addr, Handler: endpoint}
	return server.ListenAndServeTLS(certFile, keyFile)
}
