// Copyright (c) 2014 Datacratic. All rights reserved.

package rest

import (
	"fmt"
)

type router struct {
	routes   map[string]*Route
	fixed    map[string]*router
	variable map[int]*router
}

func (rt *router) Add(route *Route) *Route {
	route.Init()
	rt.add(route.Path, route)
	return route
}

func (rt *router) add(path Path, route *Route) {
	if len(path) == 0 {
		if rt.routes == nil {
			rt.routes = make(map[string]*Route)
		}

		if _, ok := rt.routes[route.Method]; ok {
			panic(fmt.Sprintf("duplicate route: %s", route.Path))
		}

		rt.routes[route.Method] = route
		return
	}

	var ok bool
	var next *router

	if path[0].IsPositional() {
		if rt.variable == nil {
			rt.variable = make(map[int]*router)
		}

		if next, ok = rt.variable[path[0].Pos]; !ok {
			next = new(router)
			rt.variable[path[0].Pos] = next
		}

	} else {
		if rt.fixed == nil {
			rt.fixed = make(map[string]*router)
		}

		if next, ok = rt.fixed[path[0].Name]; !ok {
			next = new(router)
			rt.fixed[path[0].Name] = next
		}
	}

	next.add(path[1:], route)
}

func (rt *router) Route(method, path string) (*Route, map[int]string) {
	var args map[int]string
	route := rt.route(method, splitPath(path), &args)
	return route, args
}

func (rt *router) route(method string, path []string, args *map[int]string) *Route {
	if len(path) == 0 {
		if rt.routes != nil {
			if route, ok := rt.routes[method]; ok {
				return route
			}
		}
		return nil
	}

	if rt.fixed != nil {
		if next, ok := rt.fixed[path[0]]; ok {
			return next.route(method, path[1:], args)
		}
	}

	if rt.variable != nil {
		if *args == nil {
			*args = make(map[int]string)
		}

		for pos, next := range rt.variable {
			(*args)[pos] = path[0]
			if route := next.route(method, path[1:], args); route != nil {
				return route
			}
			delete(*args, pos)
		}
	}

	return nil
}
