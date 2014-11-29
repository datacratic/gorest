// Copyright (c) 2014 Datacratic. All rights reserved.

package rest

import (
	"testing"
)

func failAdd(t *testing.T, rt *router, route *Route) {
	ret := func() (r *Route) {
		defer func() { recover() }()
		r = rt.Add(route)
		return
	}()

	if ret != nil {
		t.Errorf("FAIL: unexpected successful return: %s", route)
	}

}

func checkRouter(t *testing.T, rt *router, path, method string, expRoute *Route, expArgs ...PathItem) {
	route, args := rt.Route(method, path)
	if route != expRoute {
		t.Errorf("FAIL: routed wrong route for '%s %s' -> %s != %s",
			method, path, route, expRoute)
		return
	}

	if route == nil {
		return
	}

	if len(args) != len(expArgs) {
		t.Errorf("FAIL: args of different length for '%s %s' -> %d:%v != %d:%v",
			method, path, len(args), args, len(expArgs), expArgs)
		return
	}

	for i, exp := range expArgs {
		if i >= len(args) {
			t.Errorf("FAIL: missing arg for '%s %s' -> %s", method, path, exp)

		} else if args[i] != exp.Name {
			t.Errorf("FAIL: unexpected arg value for '%s %s' -> %s != %s",
				method, path, args[i], exp.Name)
		}
	}
}

func TestRouter(t *testing.T) {
	h0 := func() {}
	h1 := func(a int) {}
	h2 := func(a, b int) {}
	h3 := func(a, b, c int) {}

	rt := &router{}

	r00 := rt.Add(NewRoute("/", "POST", h0))
	r01 := rt.Add(NewRoute("/a", "POST", h0))
	r02 := rt.Add(NewRoute("/c", "POST", h0))
	r03 := rt.Add(NewRoute("/a/b", "POST", h0))
	r04 := rt.Add(NewRoute("/b/c", "POST", h0))
	r05 := rt.Add(NewRoute("/a/b/c", "POST", h0))
	r06 := rt.Add(NewRoute("/a/b/c", "PUT", h0))

	r10 := rt.Add(NewRoute("/:a/b/c", "POST", h1))
	r11 := rt.Add(NewRoute("/a/:b/c", "POST", h1))
	r12 := rt.Add(NewRoute("/a/b/:c", "POST", h1))
	r13 := rt.Add(NewRoute("/:a/b", "POST", h1))
	r14 := rt.Add(NewRoute("/b/:a", "POST", h2))
	r15 := rt.Add(NewRoute("/:a", "POST", h1))

	r20 := rt.Add(NewRoute("/:a/:b/c", "POST", h2))
	r21 := rt.Add(NewRoute("/:a/:b", "POST", h2))
	r22 := rt.Add(NewRoute("/:a/b/:c", "POST", h2))
	r23 := rt.Add(NewRoute("/a/:b/:c", "POST", h2))

	r30 := rt.Add(NewRoute("/:a/:b/:c", "POST", h3))

	failAdd(t, rt, NewRoute("/a/b/c", "POST", h0))
	failAdd(t, rt, NewRoute("/:a/b", "POST", h1))
	failAdd(t, rt, NewRoute("/a/b/:c", "POST", h1))

	checkRouter(t, rt, "", "POST", r00)
	checkRouter(t, rt, "/", "POST", r00)
	checkRouter(t, rt, "/a", "POST", r01)
	checkRouter(t, rt, "/c", "POST", r02)
	checkRouter(t, rt, "/a/b", "POST", r03)
	checkRouter(t, rt, "/b/c", "POST", r04)
	checkRouter(t, rt, "/a/b/c", "POST", r05)
	checkRouter(t, rt, "/a/b/c", "PUT", r06)

	checkRouter(t, rt, "/0/b/c", "POST", r10, v("0"))
	checkRouter(t, rt, "/a/1/c", "POST", r11, v("1"))
	checkRouter(t, rt, "/a/b/2", "POST", r12, v("2"))
	checkRouter(t, rt, "/3/b", "POST", r13, v("3"))
	checkRouter(t, rt, "/b/4", "POST", r14, v("4"))
	checkRouter(t, rt, "/5", "POST", r15, v("5"))

	checkRouter(t, rt, "/0/1/c", "POST", r20, v("0"), v("1"))
	checkRouter(t, rt, "/2/3", "POST", r21, v("2"), v("3"))
	checkRouter(t, rt, "/4/b/5", "POST", r22, v("4"), v("5"))
	checkRouter(t, rt, "/a/6/7", "POST", r23, v("6"), v("7"))

	checkRouter(t, rt, "/0/1/2", "POST", r30, v("0"), v("1"), v("2"))

	checkRouter(t, rt, "/a/b/c/d", "POST", nil)
	checkRouter(t, rt, "/0/b/c/d", "POST", nil)
	checkRouter(t, rt, "/a/1/c/d", "POST", nil)
	checkRouter(t, rt, "/a/b/2/d", "POST", nil)
	checkRouter(t, rt, "/0/1/2/d", "POST", nil)
	checkRouter(t, rt, "/b/c", "PUT", nil)
	checkRouter(t, rt, "/a/b/c", "DELETE", nil)
}

func BenchRouter(b *testing.B, path string) {
	h0 := func() {}
	h1 := func(a int) {}
	h2 := func(a, b int) {}
	h3 := func(a, b, c int) {}

	rt := &router{}

	rt.Add(NewRoute("/", "POST", h0))
	rt.Add(NewRoute("/a", "POST", h0))
	rt.Add(NewRoute("/c", "POST", h0))
	rt.Add(NewRoute("/a/b", "POST", h0))
	rt.Add(NewRoute("/b/c", "POST", h0))
	rt.Add(NewRoute("/a/b/c", "POST", h0))
	rt.Add(NewRoute("/a/b/c", "PUT", h0))

	rt.Add(NewRoute("/:a/b/c", "POST", h1))
	rt.Add(NewRoute("/a/:b/c", "POST", h1))
	rt.Add(NewRoute("/a/b/:c", "POST", h1))
	rt.Add(NewRoute("/:a/b", "POST", h1))
	rt.Add(NewRoute("/b/:a", "POST", h2))
	rt.Add(NewRoute("/:a", "POST", h1))

	rt.Add(NewRoute("/:a/:b/c", "POST", h2))
	rt.Add(NewRoute("/:a/:b", "POST", h2))
	rt.Add(NewRoute("/:a/b/:c", "POST", h2))
	rt.Add(NewRoute("/a/:b/:c", "POST", h2))

	rt.Add(NewRoute("/:a/:b/:c", "POST", h3))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rt.Route("POST", path)
	}
}

func BenchmarkRouterRoot(b *testing.B) {
	BenchRouter(b, "")
}

func BenchmarkRouter1Fix(b *testing.B) {
	BenchRouter(b, "/a")
}

func BenchmarkRouter2Fix(b *testing.B) {
	BenchRouter(b, "/a/b")
}

func BenchmarkRouter3Fix(b *testing.B) {
	BenchRouter(b, "/a/b/c")
}

func BenchmarkRouter1Var(b *testing.B) {
	BenchRouter(b, "/1/b/c")
}

func BenchmarkRouter2Var(b *testing.B) {
	BenchRouter(b, "/1/2/c")
}

func BenchmarkRouter3Var(b *testing.B) {
	BenchRouter(b, "/1/2/3")
}

func BenchmarkRouterUnknownShallow(b *testing.B) {
	BenchRouter(b, "/d")
}

func BenchmarkRouterUnknownDeep(b *testing.B) {
	BenchRouter(b, "/a/b/c/d")
}

func BenchmarkRouterUnknownVariable(b *testing.B) {
	BenchRouter(b, "/1/2/3/4")
}
