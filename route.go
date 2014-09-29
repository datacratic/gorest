// Copyright (c) 2014 Datacratic. All rights reserved.

package rest

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"sync"
)

// Routable is used to detect objects that are routable by an Endpoint.
type Routable interface {

	// RESTRoutes returns a list of Route that can be used for routing REST
	// requests.
	RESTRoutes() Routes
}

// Routes is convenience type that represents a list of Route objects.
type Routes []*Route

// Route associates a handler which should be invoked for a given HTTP method
// and templated path.
type Route struct {

	// Method represents the HTTP verb required by this route.
	Method string

	// Path is the templated path required by this route. See Handler for the
	// rules related to path.
	Path Path

	// Handler is a function to be invoked whenever for a given HTTP method and
	// templated path.
	//
	// The function can only return at most 2 values where one will be an error
	// object and the other will be the body of the HTTP response.
	//
	// The function can have N + 1 arguments where N is the number of positional
	// parameters in Path and where the +1 is the argument that will filled from
	// the body of the HTTP request.
	//
	// No restrictions are applied on the order of the arguments, return values
	// so long as all the arguments can be filled in from a given HTTP request.
	//
	// If any of the previous rules are broken, Route will panic when Init is
	// called.
	Handler interface{}

	initialize sync.Once

	handler     reflect.Value
	handlerType reflect.Type

	inBody   int
	outBody  int
	outError int
}

// NewRoute creates and initializes a new Route from the method, path and
// handler.
func NewRoute(method, path string, handler interface{}) *Route {
	route := &Route{
		Method:  method,
		Path:    NewPath(path),
		Handler: handler,
	}
	route.Init()
	return route
}

// Init initializes the object.
func (route *Route) Init() {
	route.initialize.Do(route.init)
}

func (route *Route) init() {
	route.handler = reflect.ValueOf(route.Handler)
	route.handlerType = route.handler.Type()

	if route.handlerType.Kind() != reflect.Func {
		panic(fmt.Sprintf("invalid handler type for route { %s %s }: got '%s' expected '%s'",
			route.Method, route.Path, route.handlerType.Kind(), reflect.Func))
	}

	args := make(map[int]struct{})
	for i := 0; i < route.handlerType.NumIn(); i++ {
		args[i] = struct{}{}
	}

	for _, item := range route.Path {
		if !item.IsPositional() {
			continue
		}

		if _, ok := args[item.Pos]; !ok {
			panic(fmt.Sprintf(
				"out of bound or duplicate positional argument for route { %s %s }: '%s' -> %d >= %d",
				route.Method, route.Path, item, item.Pos, route.handlerType.NumIn()))
		}

		delete(args, item.Pos)
	}

	if len(args) > 1 {
		panic(fmt.Sprintf("can only have one non-positional handler argument for route %s", route))
	}

	route.inBody = -1
	if len(args) > 0 {
		for pos := range args {
			route.inBody = pos
		}
	}

	if route.handlerType.NumOut() > 2 {
		panic(fmt.Sprintf("too many return arguments for route %s", route))
	}

	route.outBody = -1
	route.outError = -1

	for i := 0; i < route.handlerType.NumOut(); i++ {
		errorType := reflect.TypeOf((*error)(nil)).Elem()

		if out := route.handlerType.Out(i); out == errorType {
			if route.outError >= 0 {
				panic(fmt.Sprintf("too many error return for route %s", route))
			}
			route.outError = i

		} else {
			if route.outBody >= 0 {
				panic(fmt.Sprintf("too many normal return for route %s", route))
			}
			route.outBody = i
		}
	}
}

func (route *Route) parseArg(data string, value reflect.Value) (err error) {
	switch value.Kind() {

	case reflect.String:
		value.SetString(data)

	case reflect.Bool:
		var b bool
		if b, err = strconv.ParseBool(data); err == nil {
			value.SetBool(b)
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var i int64
		if i, err = strconv.ParseInt(data, 10, value.Type().Bits()); err == nil {
			value.SetInt(i)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var u uint64
		if u, err = strconv.ParseUint(data, 10, value.Type().Bits()); err == nil {
			value.SetUint(u)
		}

	case reflect.Float32, reflect.Float64:
		var f float64
		if f, err = strconv.ParseFloat(data, value.Type().Bits()); err == nil {
			value.SetFloat(f)
		}

	default:
		err = fmt.Errorf("unsupported argument type for route '%s %s': %s",
			route.Method, route.Path, value.Kind())
	}

	return
}

func (route *Route) invoke(args map[int]string, body []byte) ([]byte, ErrorType, error) {
	var in []reflect.Value
	for i := 0; i < route.handlerType.NumIn(); i++ {
		arg := reflect.New(route.handlerType.In(i))

		var err error

		if data, ok := args[i]; ok {
			err = route.parseArg(data, arg.Elem())
		} else {
			err = json.Unmarshal(body, arg.Interface())
		}

		if err != nil {
			return nil, UnmarshalError, err
		}

		in = append(in, arg.Elem())
	}

	out := route.handler.Call(in)

	if route.outError >= 0 && !out[route.outError].IsNil() {
		return nil, HandlerError, out[route.outError].Interface().(error)
	}

	var ret []byte

	if route.outBody >= 0 {
		var err error
		ret, err = json.Marshal(out[route.outBody].Interface())
		if err != nil {
			return nil, MarshalError, err
		}
	}

	return ret, NoError, nil
}

// String returns a string represenation of the object suitable for debugging.
func (route *Route) String() string {
	return fmt.Sprintf("{ %s %s }", route.Method, route.Path)
}
