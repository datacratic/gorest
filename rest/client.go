// Copyright (c) 2014 Datacratic. All rights reserved.

package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Client is a convenience REST client which simplifies the creation of REST
// requests.
type Client struct {
	*http.Client

	// Host is the address of the remote REST endpoint where requests should be
	// sent to.
	Host string

	// Root is a prefix that will be preprended to all path requests created by
	// this client.
	Root string

	// Header is a list of HTTP requests that will be added to every requests
	// originating from this client.
	Header http.Header

	// Limit sets a hard limit on the number of concurrent requests. If not set
	// then no limits are imposed.
	Limit uint

	initialize sync.Once

	limit chan struct{}
}

// NewRequest creates a new Request object for the given HTTP method.
func (client *Client) NewRequest(method string) *Request {
	client.initialize.Do(func() {
		if client.Client == nil {
			client.Client = http.DefaultClient
		}

		if client.Limit > 0 {
			client.limit = make(chan struct{}, client.Limit)

			for i := uint(0); i < client.Limit; i++ {
				client.end()
			}
		}

	})

	headers := make(http.Header)
	if client.Header != nil {
		for key, val := range client.Header {
			headers[key] = val
		}
	}

	return &Request{
		REST:   client,
		Client: client.Client,
		Host:   client.Host,
		Method: method,
		Root:   client.Root,
		Header: headers,
	}
}

func (client *Client) begin() {
	if client.limit != nil {
		<-client.limit
	}
}

func (client *Client) end() {
	if client.limit != nil {
		client.limit <- struct{}{}
	}
}

// Request is used to gradually construct REST requests and send them to a
// remote endpoint. Request object should be created via the NewRequest function
// or the Client.NewRequest function and filled in via the SetXxx and AddXxx
// functions. Finally the request is sent via the Send function which returns a
// Response object.
type Request struct {

	// REST is the client that originated the request. Can be nil.
	REST *Client

	// Client is the http.Client used to send the request. Defaults to
	// http.DefaultClient and can changed via the SetClient method.
	Client *http.Client

	// Host is the address of the remote REST endpoint where requests should be
	// sent to.
	Host string

	// Root is a prefix that will be preprended to all path requests created by
	// this client.
	Root string

	// Path is the absolute path where the Request should be routed to on the
	// remote endpoint. Can be changed via the SetPath method.
	Path string

	// Query contains the url parameters and values that will be sent with the
	// request. Paramerters can be added via the AddParam method.
	Query url.Values

	// Method is the HTTP verb used for the HTTP request.
	Method string

	// Header contains all the headers to be added to the HTTP request. Can be
	// changed via the AddHeader method.
	Header http.Header

	// Body is the JSON serialized body of the HTTP request. Can be set via the
	// SetBody method.
	Body []byte

	HTTP *http.Request

	err *Error
}

// NewRequest creates a new Request object to be sent to the given host using
// the given HTTP verb.
func NewRequest(host, method string) *Request {
	return &Request{
		Host:   host,
		Method: method,
		Client: http.DefaultClient,
	}
}

// SetHost selects the host where the request will be sent.
func (req *Request) SetHost(host string) *Request {
	req.Host = host
	return req
}

// SetClient selects the http.Client to be used to execute the requests.
func (req *Request) SetClient(client *http.Client) *Request {
	req.Client = client
	return req
}

// SetPath formats and sets the path where the request will be routed to. Note
// that the root is prefixed to the path before formatting the string.
func (req *Request) SetPath(path string, args ...interface{}) *Request {
	req.Path = fmt.Sprintf(JoinPath(req.Root, path), args...)
	return req
}

// AddParam adds a parameter to the query string.
func (req *Request) AddParam(key, value string) *Request {
	if req.Query == nil {
		req.Query = url.Values{}
	}
	req.Query.Add(key, value)
	return req
}

// AddHeader adds the given header to the request.
func (req *Request) AddHeader(key, value string) *Request {
	if req.Header == nil {
		req.Header = make(http.Header)
	}

	req.Header.Add(key, value)
	return req
}

// SetBody marshals the given objects and sets it as the body of the
// request. The Content-Length header will be automatically set.
func (req *Request) SetBody(obj interface{}) *Request {
	var err error
	if req.Body, err = json.Marshal(obj); err == nil {
		req.AddHeader("Content-Length", strconv.Itoa(len(req.Body)))

	} else {
		req.err = &Error{MarshalError, err}
	}

	return req
}

func (req *Request) SetRawBody(obj json.RawMessage) *Request {
	req.Body = obj
	req.AddHeader("Content-Length", strconv.Itoa(len(obj)))
	return req
}

// Send attempts to send the request to the remote endpoint and returns a
// Response which contains the result.
func (req *Request) Send() *Response {
	t0 := time.Now()

	if len(req.Path) == 0 {
		req.Path = req.Root
	}

	resp := &Response{Request: req, Error: req.err}

	if resp.Error == nil {
		if req.REST != nil {
			req.REST.begin()
		}

		req.send(resp)

		if req.REST != nil {
			req.REST.end()
		}
	}

	resp.Latency = time.Since(t0)
	return resp
}

func (req *Request) send(resp *Response) {
	var reader io.Reader
	if len(req.Body) > 0 {
		reader = bytes.NewReader(req.Body)
	}

	urlS := strings.TrimRight(req.Host, "/") + req.Path

	if req.Query != nil {
		urlS += "?" + req.Query.Encode()
	}

	var err error

	if req.HTTP, err = http.NewRequest(req.Method, urlS, reader); err != nil {
		resp.Error = &Error{NewRequestError, err}
		return
	}

	req.AddHeader("Content-Type", "application/json")
	req.HTTP.Header = req.Header

	httpResp, err := req.Client.Do(req.HTTP)
	if err != nil {
		if err2, ok := err.(*url.Error); ok {
			if err3, ok := err2.Err.(net.Error); ok {
				if err3.Timeout() {
					resp.Error = &Error{TimeoutError, err}
					return
				}
			}
		}
		resp.Error = &Error{SendRequestError, err}
		return
	}

	resp.Code = httpResp.StatusCode
	resp.Header = httpResp.Header

	if resp.Body, err = ioutil.ReadAll(httpResp.Body); err != nil {
		resp.Error = &Error{ReadBodyError, err}
	}

	httpResp.Body.Close()
	return
}

// Response holds the result of a sent REST request. The response should be read
// via the GetBody method which checks the various fields to detect errors.
type Response struct {

	// Request is the request that originated the response.
	Request *Request

	// Code is the http status code returned by the endpoint.
	Code int

	// Header holds the headers of the HTTP response.
	Header http.Header

	// Body holds the raw unmarshalled body of the HTTP response. GetBody can be
	// used to unmarshal the body.
	Body []byte

	// Error is set if an error occured while sending the request.
	Error *Error

	// Latency indicates how long the request round-trip took.
	Latency time.Duration
}

// GetBody checks the various fields of the response for errors and unmarshals
// the response body if the given object is not nil. If an error is detected,
// the error type and error will be returned instead.
func (resp *Response) GetBody(obj interface{}) (err *Error) {
	if resp.Error != nil {
		err = resp.Error

	} else if resp.Code == http.StatusNotFound {
		err = &Error{UnknownRoute, errors.New(string(resp.Body))}

	} else if resp.Code >= 400 {
		err = &Error{EndpointError, errors.New(string(resp.Body))}

	} else if resp.Code < 200 && resp.Code >= 300 {
		err = ErrorFmt(UnexpectedStatusCode, "unexpected status code: %d", resp.Code)

	} else if resp.Code == http.StatusNoContent {
		if obj == nil {
			return
		}
		err = ErrorFmt(UnexpectedStatusCode, "unexpected status code: 204")

	} else if contentType := resp.Header.Get("Content-Type"); len(resp.Body) > 0 && contentType != "application/json" {
		err = ErrorFmt(UnsupportedContentType, "unsupported content-type: '%s' != 'application/json'", contentType)

	} else if obj == nil {
		return

	} else if jsonErr := json.Unmarshal(resp.Body, obj); jsonErr != nil {
		err = &Error{UnmarshalError, jsonErr}
	}

	return
}
