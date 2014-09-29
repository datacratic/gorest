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

func checkRouter(t *testing.T, rt *router, method, path string, expRoute *Route, expArgs ...PathItem) {
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

	for _, exp := range expArgs {
		arg, ok := args[exp.Pos]
		if !ok {
			t.Errorf("FAIL: missing arg for '%s %s' -> %s", method, path, exp)

		} else if arg != exp.Name {
			t.Errorf("FAIL: unexpected arg value for '%s %s' and pos %d -> %s != %s",
				method, path, exp.Pos, arg, exp.Name)
		}
	}
}

func TestRouter(t *testing.T) {
	h0 := func() {}
	h1 := func(a int) {}
	h2 := func(a, b int) {}
	h3 := func(a, b, c int) {}

	rt := &router{}

	r00 := rt.Add(NewRoute("POST", "/", h0))
	r01 := rt.Add(NewRoute("POST", "/a", h0))
	r02 := rt.Add(NewRoute("POST", "/c", h0))
	r03 := rt.Add(NewRoute("POST", "/a/b", h0))
	r04 := rt.Add(NewRoute("POST", "/b/c", h0))
	r05 := rt.Add(NewRoute("POST", "/a/b/c", h0))
	r06 := rt.Add(NewRoute("PUT", "/a/b/c", h0))

	r10 := rt.Add(NewRoute("POST", "/{0:a}/b/c", h1))
	r11 := rt.Add(NewRoute("POST", "/a/{0:b}/c", h1))
	r12 := rt.Add(NewRoute("POST", "/a/b/{0:c}", h1))
	r13 := rt.Add(NewRoute("POST", "/{0:a}/b", h1))
	r14 := rt.Add(NewRoute("POST", "/b/{1:a}", h2))
	r15 := rt.Add(NewRoute("POST", "/{0:a}", h1))

	r20 := rt.Add(NewRoute("POST", "/{0:a}/{1:b}/c", h2))
	r21 := rt.Add(NewRoute("POST", "/{1:a}/{0:b}", h2))
	r22 := rt.Add(NewRoute("POST", "/{0:a}/b/{1:c}", h2))
	r23 := rt.Add(NewRoute("POST", "/a/{1:b}/{0:c}", h2))

	r30 := rt.Add(NewRoute("POST", "/{0:a}/{1:b}/{2:c}", h3))

	failAdd(t, rt, NewRoute("POST", "/a/b/c", h0))
	failAdd(t, rt, NewRoute("POST", "/{0:a}/b", h1))
	failAdd(t, rt, NewRoute("POST", "/a/b/{0:c}", h1))

	checkRouter(t, rt, "POST", "", r00)
	checkRouter(t, rt, "POST", "/", r00)
	checkRouter(t, rt, "POST", "/a", r01)
	checkRouter(t, rt, "POST", "/c", r02)
	checkRouter(t, rt, "POST", "/a/b", r03)
	checkRouter(t, rt, "POST", "/b/c", r04)
	checkRouter(t, rt, "POST", "/a/b/c", r05)
	checkRouter(t, rt, "PUT", "/a/b/c", r06)

	checkRouter(t, rt, "POST", "/0/b/c", r10, p(0, "0"))
	checkRouter(t, rt, "POST", "/a/1/c", r11, p(0, "1"))
	checkRouter(t, rt, "POST", "/a/b/2", r12, p(0, "2"))
	checkRouter(t, rt, "POST", "/3/b", r13, p(0, "3"))
	checkRouter(t, rt, "POST", "/b/4", r14, p(1, "4"))
	checkRouter(t, rt, "POST", "/5", r15, p(0, "5"))

	checkRouter(t, rt, "POST", "/0/1/c", r20, p(0, "0"), p(1, "1"))
	checkRouter(t, rt, "POST", "/2/3", r21, p(1, "2"), p(0, "3"))
	checkRouter(t, rt, "POST", "/4/b/5", r22, p(0, "4"), p(1, "5"))
	checkRouter(t, rt, "POST", "/a/6/7", r23, p(1, "6"), p(0, "7"))

	checkRouter(t, rt, "POST", "/0/1/2", r30, p(0, "0"), p(1, "1"), p(2, "2"))

	checkRouter(t, rt, "POST", "/a/b/c/d", nil)
	checkRouter(t, rt, "POST", "/0/b/c/d", nil)
	checkRouter(t, rt, "POST", "/a/1/c/d", nil)
	checkRouter(t, rt, "POST", "/a/b/2/d", nil)
	checkRouter(t, rt, "POST", "/0/1/2/d", nil)
	checkRouter(t, rt, "PUT", "/b/c", nil)
	checkRouter(t, rt, "DELETE", "/a/b/c", nil)
}

func BenchRouter(b *testing.B, path string) {
	h0 := func() {}
	h1 := func(a int) {}
	h2 := func(a, b int) {}
	h3 := func(a, b, c int) {}

	rt := &router{}

	rt.Add(NewRoute("POST", "/", h0))
	rt.Add(NewRoute("POST", "/a", h0))
	rt.Add(NewRoute("POST", "/c", h0))
	rt.Add(NewRoute("POST", "/a/b", h0))
	rt.Add(NewRoute("POST", "/b/c", h0))
	rt.Add(NewRoute("POST", "/a/b/c", h0))
	rt.Add(NewRoute("PUT", "/a/b/c", h0))

	rt.Add(NewRoute("POST", "/{0:a}/b/c", h1))
	rt.Add(NewRoute("POST", "/a/{0:b}/c", h1))
	rt.Add(NewRoute("POST", "/a/b/{0:c}", h1))
	rt.Add(NewRoute("POST", "/{0:a}/b", h1))
	rt.Add(NewRoute("POST", "/b/{1:a}", h2))
	rt.Add(NewRoute("POST", "/{0:a}", h1))

	rt.Add(NewRoute("POST", "/{0:a}/{1:b}/c", h2))
	rt.Add(NewRoute("POST", "/{1:a}/{0:b}", h2))
	rt.Add(NewRoute("POST", "/{0:a}/b/{1:c}", h2))
	rt.Add(NewRoute("POST", "/a/{1:b}/{0:c}", h2))

	rt.Add(NewRoute("POST", "/{0:a}/{1:b}/{2:c}", h3))

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
