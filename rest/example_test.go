// Copyright (c) 2014 Datacratic. All rights reserved.

package rest_test

import (
	"github.com/datacratic/gorest/rest"

	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

// Our service that we want to expose through a REST api.
type PingService struct{}

func (*PingService) Ping(tick int) int { return tick }
func (*PingService) PingError() error  { return errors.New("BOOM") }
func (*PingService) PingDelay() int    { time.Sleep(5 * time.Second); return 1 }

// RESTRoutes implements the rest.Routable interface which is used by
// rest.Endpoint to route incoming HTTP request to the appropriate handler.
func (svc *PingService) RESTRoutes() rest.Routes {

	return rest.Routes{

		// Route with no arguments provided in the path means that the argument
		// required by Ping will be taken from the body of the HTTP request. The
		// body is assumed to be JSON and will automatically be unmarshalled.
		//
		// The function's return value will be automatically serialized to JSON
		// using the encoding.json package and used as the body of the HTTP
		// response sent to the client. If nothing or nil is returned then an
		// HTTP 204 response code is returned instead.
		rest.NewRoute("/ping", "POST", svc.Ping),

		// Path components starting with ':' indicates that the first argument
		// of the Ping function will be fed from the value found where the
		// placeholder is in the path.
		//
		// Path arguments are assumed to be basic types (string, bool, int
		// floats) and are automatically converted to match the function
		// argument.
		rest.NewRoute("/ping/:tick", "PUT", svc.Ping),

		// We can also return errors which will automatically be detected and
		// converted to a HTTP 400 return code to the user. The HTTP return code
		// can be changed by returning a rest.CodedError.
		//
		// The body is returned as a simple string containing the string
		// representation of the error object.
		rest.NewRoute("/ping/error", "POST", svc.PingError),

		// Used for testing the Timeout() feature of a request.
		rest.NewRoute("/ping/delay", "GET", svc.PingDelay),
	}
}

func ExamplePing() {

	// Add our service that implements the rest.Routable interface to our
	// endpoint. We can also add simple lambda functions to our endpoint.
	rest.AddService(new(PingService))
	rest.AddRoute("/simple", "POST", func(tick int) int { return tick })

	// The endpoint is started like any other http.Server.
	go rest.ListenAndServe(":11263", nil)

	time.Sleep(30 * time.Millisecond)

	// The rest package also provides a way to query a REST endpoint by
	// incrementally building a REST request. The body of the query can be set
	// via the SetBody() function which will serialize the object to JSON using
	// the encoding.json package.
	simpleResp := rest.NewRequest("http://localhost:11263", "POST").
		SetPath("/simple").
		SetBody(123).
		Send()

	// The rest.Response object contains the result of the REST request and can
	// be used to deserialize the JSON body into the provided object using the
	// encoding.json package.
	var tick int
	if err := simpleResp.GetBody(&tick); err != nil {
		panic("Whoops! " + fmt.Sprint(err))
	}
	fmt.Println("ping-simple:", tick)

	// The REST client can also be customized by manually creating a rest.Client
	// which can then be used to create REST requests. rest.Client embdeds an
	// http.Client struct which can be used to customize the HTTP requests.
	client := &rest.Client{Host: "http://localhost:11263", Root: "/ping"}
	clientResp := client.NewRequest("PUT").
		SetPath("%d", 321).
		AddParam("test", "321").
		AddParam("test2", "3212").
		Send()

	if err := clientResp.GetBody(&tick); err != nil {
		panic("Whoops!")
	}
	fmt.Println("ping-client:", tick)

	// The REST client can check if a response timed out.
	dialer := net.Dialer{Timeout: 500 * time.Millisecond}
	clientTime := &rest.Client{
		Host: "http://localhost:11263",
		Root: "/ping",
		Client: &http.Client{
			Transport: &http.Transport{Dial: dialer.Dial},
			Timeout:   500 * time.Millisecond,
		},
	}
	timeoutResp := clientTime.NewRequest("GET").
		SetPath("delay").
		Send()

	if err := timeoutResp.GetBody(&tick); err == nil {
		panic("Whoops! Should time out" + fmt.Sprint(tick))
	}

	// Output:
	// ping-simple: 123
	// ping-client: 321
}
