// Copyright (c) 2014 Datacratic. All rights reserved.

package rest

import (
	"log"
)

type router struct {
	routes   map[string]*Route
	fixed    map[string]*router
	variable *router
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
			log.Panicf("duplicate route: %s", route.Path)
		}

		rt.routes[route.Method] = route
		return
	}

	var ok bool
	var next *router

	if path[0].IsArg {
		if rt.variable == nil {
			rt.variable = new(router)
		}
		next = rt.variable

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

func (rt *router) Route(method, path string) (*Route, []string) {
	var args []string
	return rt.route(method, SplitPath(path), args)
}

func (rt *router) route(method string, path []string, args []string) (*Route, []string) {
	if len(path) == 0 {
		if rt.routes != nil {
			if route, ok := rt.routes[method]; ok {
				return route, args
			}
		}
		return nil, args
	}

	if rt.fixed != nil {
		if next, ok := rt.fixed[path[0]]; ok {
			return next.route(method, path[1:], args)
		}
	}

	if rt.variable != nil {
		args = append(args, path[0])
		return rt.variable.route(method, path[1:], args)
	}

	return nil, args
}

func (rt *router) PrintRoutes(routes []*Route) []*Route {
	if rt.routes != nil {
		for _, route := range rt.routes {
			routes = append(routes, route)
		}
	}
	if rt.fixed != nil {
		for _, r := range rt.fixed {
			routes = r.PrintRoutes(routes)
		}
	}
	if rt.variable != nil {
		routes = rt.variable.PrintRoutes(routes)
	}
	return routes
}
