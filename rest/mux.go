// Copyright (c) 2014 Datacratic. All rights reserved.

package rest

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

//go:generate go run templates/include_templates.go

// Mux routes incoming bid requests to the registered routes. Implements
// the http.Handler interface.
//
// The mux currently only supports JSON content-type for regular message
// and text/plain for error messages.
type Mux struct {

	// Root is the path prefix of all the routes to be matched by this
	// mux. Must be set before calling Init and can't be changed
	// afterwards.
	Root string

	// ErrorFunc is called for all errors that passes through this mux. The
	// returned value will overwrite the current error and will be returned to
	// the client instead.. If it's return value is a rest.CodedError then the
	// status code of the error object will be used.
	ErrorFunc func(ErrorType, error) error

	DefaultHandler http.Handler

	initialize sync.Once

	router router
}

// Init initializes the object.
func (mux *Mux) Init() {
	mux.initialize.Do(mux.init)
}

func (mux *Mux) init() {
	if len(mux.Root) == 0 {
		mux.Root = "/"
	} else {
		mux.Root = "/" + strings.Trim(mux.Root, "/")
	}

	if mux.DefaultHandler == nil {
		mux.DefaultHandler = http.DefaultServeMux
	}
}

// AddRoute adds all the given routes to the mux.
func (mux *Mux) AddRoute(routes ...*Route) {
	mux.Init()

	for _, route := range routes {
		mux.router.Add(route)
	}
}

// AddService adds all the routes returned by the Routable objects to the mux.
func (mux *Mux) AddService(routables ...Routable) {
	for _, routable := range routables {
		mux.AddRoute(routable.RESTRoutes()...)
	}
}

func (mux *Mux) route(method, path string) (*Route, []string, error) {
	if strings.HasPrefix(path, mux.Root) {
		sub := path[len(mux.Root):]
		if route, args := mux.router.Route(method, sub); route != nil {
			return route, args, nil
		}
	}

	return nil, nil, fmt.Errorf("unknown path: '%s'", path)
}

func (mux *Mux) respondError(writer http.ResponseWriter, errType ErrorType, code int, err error) {
	if mux.ErrorFunc != nil {
		err = mux.ErrorFunc(errType, err)
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
func (mux *Mux) ServeHTTP(writer http.ResponseWriter, httpReq *http.Request) {
	mux.Init()

	if httpReq.URL.Path == "/documentation" {
		funcMap := make(template.FuncMap)
		funcMap["Split"] = strings.Split
		funcMap["Contains"] = strings.Contains
		t, err := template.New("documentation").Funcs(funcMap).Parse(documentation)
		if err != nil {
			mux.respondError(writer, "html-template-error", http.StatusBadRequest, err)
			return
		}

		routes := mux.router.PrintRoutes(make(Routes, 0))
		sort.Sort(routes)

		page := struct {
			Host   string
			Routes Routes
		}{
			httpReq.Host,
			routes,
		}
		t.Execute(writer, page)
		return
	}

	route, args, err := mux.route(httpReq.Method, httpReq.URL.Path)
	if err != nil {
		mux.DefaultHandler.ServeHTTP(writer, httpReq)
		return
	}

	if httpReq.Method != "GET" {
		if contentType := httpReq.Header.Get("Content-Type"); contentType != "application/json" {
			err := fmt.Errorf("unsupported content type: got '%s' expected 'application/json'", contentType)
			mux.respondError(writer, UnsupportedContentType, http.StatusBadRequest, err)
			return
		}
	}

	var body []byte
	if contentEncoding := httpReq.Header.Get("Content-Encoding"); contentEncoding == "gzip" {
		gz, err := gzip.NewReader(httpReq.Body)
		defer gz.Close()
		if err != nil {
			err := fmt.Errorf("decoding gzip content failed: %s", err)
			mux.respondError(writer, GzipError, http.StatusBadRequest, err)
			return
		}
		body, err = ioutil.ReadAll(gz)
		if err != nil {
			err := fmt.Errorf("decoding gzip content failed: %s", err)
			mux.respondError(writer, GzipError, http.StatusBadRequest, err)
			return
		}
	} else {
		var err error
		body, err = ioutil.ReadAll(httpReq.Body)
		if err != nil {
			mux.respondError(writer, ReadBodyError, http.StatusBadRequest, err)
			return
		}
	}

	resp, restError := route.invoke(args, body)
	if restError != nil {
		mux.respondError(writer, restError.Type, http.StatusBadRequest, restError.Sub)
		return
	}

	if len(resp) == 0 {
		writer.WriteHeader(http.StatusNoContent)
	} else {
		header := writer.Header()

		if route.GzipLevel != 0 {
			var body bytes.Buffer
			gz, _ := gzip.NewWriterLevel(&body, route.GzipLevel)
			_, err := gz.Write(resp)
			if err != nil {
				err := fmt.Errorf("decoding gzip content failed: %s", err)
				mux.respondError(writer, GzipError, http.StatusBadRequest, err)
				return
			}
			gz.Close()
			resp = body.Bytes()
			header.Set("Content-Encoding", "gzip")
		}

		header.Set("Content-Type", "application/json")
		header.Set("Content-Length", strconv.FormatInt(int64(len(resp)), 10))
		writer.Write(resp)
	}
}
