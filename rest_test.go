// Copyright (c) 2014 Datacratic. All rights reserved.

package rest_test

import (
	"github.com/datacratic/gortbkit/deps/gorest"

	"errors"
	"fmt"
)

//
type PingService struct{}

func (*PingService) Ping(tick int) int { return tick }
func (*PingService) PingError() error  { return errors.New("BOOM") }

func (svc *PingService) RESTRoutes() rest.Routes {

	// \todo: go vet chokes if we put rest.Routes instead of []*rest.Route.
	return []*rest.Route{

		// Route with no arguments provided in the path means that the argument
		// required by Ping will be taken from the body of the HTTP request. The
		// body is assumed to be JSON and will automatically be unmarshalled.
		rest.NewRoute("POST", "/ping", svc.Ping),

		// {0:tick} indicates that the first argument of the Ping function will
		// be fed from the value found where the placeholder is in the path.
		//
		// Path arguments are assumed to be basic types (string, bool, int
		// floats) and are automatically converted to match the function
		// argument.
		rest.NewRoute("PUT", "/ping/:tick", svc.Ping),

		// We can also return errors which will automatically be detected and
		// converted to a HTTP 400 return code to the user. The HTTP return code
		// can be changed by returning a rest.CodedError.
		//
		//The body is returned as a simple string containing the string
		//representation of the error object.
		rest.NewRoute("POST", "/ping/error", svc.PingError),
	}
}

func ExamplePing() {
	var tick int

	endpoint := rest.TestEndpoint{
		Endpoint: rest.Endpoint{Root: "/v1"},
	}

	endpoint.AddRoutable(&PingService{})
	endpoint.AddRoute(rest.NewRoute("POST", "/simple", func(tick int) int { return tick }))

	endpoint.ListenAndServe()

	simpleResp := rest.NewRequest(endpoint.RootedURL(), "POST").
		SetPath("/simple").
		SetBody(123).
		Send()

	if err := simpleResp.GetBody(&tick); err != nil {
		panic("Whoops!")
	}
	fmt.Println("ping-simple:", tick)

	client := &rest.Client{Host: endpoint.URL(), Root: "/v1/ping"}
	clientResp := client.NewRequest("PUT").
		SetPath("%d", 321).
		Send()

	if err := clientResp.GetBody(&tick); err != nil {
		panic("Whoops!")
	}
	fmt.Println("ping-client:", tick)

	// Output:
	// ping-simple: 123
	// ping-client: 321
}
