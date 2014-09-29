// Copyright (c) 2014 Datacratic. All rights reserved.

package rest

import (
	"bytes"
	"fmt"
	"testing"
)

func printPath(path ...PathItem) string {
	buffer := new(bytes.Buffer)
	buffer.WriteString("/")

	for _, item := range path {
		buffer.WriteString(item.String())
		buffer.WriteString("/")
	}

	return buffer.String()
}

func f(name string) PathItem {
	return PathItem{name, -1}
}

func p(pos int, name string) PathItem {
	return PathItem{name, pos}
}

func checkRoute(t *testing.T, handler interface{}, path string, exp ...PathItem) (route *Route) {
	route = NewRoute("POST", path, handler)

	if len(exp) != len(route.Path) {
		t.Errorf("FAIL: path length mismatch: %d:%s != %d:%s",
			len(exp), printPath(exp...), len(route.Path), printPath(route.Path...))
		return
	}

	for i, comp := range route.Path {
		if comp.Name != exp[i].Name || comp.Pos != exp[i].Pos {
			t.Errorf("FAIL: PathItem mismatch: %d -> %s:%s != %s:%s",
				i, exp[i], printPath(exp...), comp, printPath(route.Path...))
		}
	}

	return
}

func failRoute(t *testing.T, handler interface{}, path string) {
	route := func() (route *Route) {
		defer func() { recover() }()
		route = NewRoute("POST", path, handler)
		return
	}()

	if route != nil {
		t.Errorf("FAIL: unexpected successful return: %s -> %s", path, printPath(route.Path...))
	}
}

type T struct {
	Value int `json:"val"`
}

func TestRouteInit(t *testing.T) {

	hNoop := func() {}

	checkRoute(t, hNoop, "")
	checkRoute(t, hNoop, "/")
	checkRoute(t, hNoop, "///")

	checkRoute(t, hNoop, "a", f("a"))
	checkRoute(t, hNoop, "/a", f("a"))
	checkRoute(t, hNoop, "a/", f("a"))
	checkRoute(t, hNoop, "/a/", f("a"))

	checkRoute(t, hNoop, "a/b", f("a"), f("b"))
	checkRoute(t, hNoop, "/a/b/", f("a"), f("b"))
	checkRoute(t, hNoop, "a/b/c", f("a"), f("b"), f("c"))

	h1Arg := func(int) int { return 0 }

	checkRoute(t, h1Arg, "")
	checkRoute(t, h1Arg, "/a", f("a"))
	checkRoute(t, h1Arg, "/{0:a}", p(0, "a"))
	checkRoute(t, h1Arg, "a/{0:b}/c", f("a"), p(0, "b"), f("c"))

	failRoute(t, h1Arg, "{1:a}")
	failRoute(t, h1Arg, "{0:a}/{0:b}")
	failRoute(t, h1Arg, "{0:a}/{1:b}")

	h2Arg := func(T, int) (t T, err error) { return }

	failRoute(t, h2Arg, "")
	failRoute(t, h2Arg, "a")
	failRoute(t, h2Arg, "a/b")

	checkRoute(t, h2Arg, "{0:a}", p(0, "a"))
	checkRoute(t, h2Arg, "{0:a}/{1:b}", p(0, "a"), p(1, "b"))
	checkRoute(t, h2Arg, "{1:a}/{0:b}", p(1, "a"), p(0, "b"))
	checkRoute(t, h2Arg, "c/{1:a}/d/{0:b}/e", f("c"), p(1, "a"), f("d"), p(0, "b"), f("e"))

	failRoute(t, h2Arg, "{2:a}")
	failRoute(t, h2Arg, "{0:a}/{1:b}/{2:c}")

	failRoute(t, T{}, "")
	failRoute(t, &T{}, "")

	failRoute(t, func() (e0 error, e1 error) { return }, "")
	failRoute(t, func() (i0 int, i1 int) { return }, "")
	failRoute(t, func() (i0 int, i1 int, i2 int) { return }, "")
}

func checkInvoke(t *testing.T, route *Route, exp string, body string, args ...PathItem) {
	m := make(map[int]string)
	for _, arg := range args {
		m[arg.Pos] = arg.Name
	}

	ret, errT, err := route.invoke(m, []byte(body))

	if err != nil {
		t.Errorf("FAIL%s: unexpected error '%s','%s' -> %d:%s",
			route, body, printPath(args...), errT, err)
		return
	}

	if string(ret) != exp {
		t.Errorf("FAIL%s: return mismatch '%s','%s' -> %s != %s",
			route, body, printPath(args...), string(ret), exp)
		return
	}
}

func failInvoke(t *testing.T, route *Route, exp ErrorType, body string, args ...PathItem) {
	m := make(map[int]string)
	for _, arg := range args {
		m[arg.Pos] = arg.Name
	}

	ret, errT, err := route.invoke(m, []byte(body))

	if err == nil {
		t.Errorf("FAIL%s: unexpected return '%s','%s' -> %s",
			route, body, printPath(args...), string(ret))
		return
	}

	if errT != exp {
		t.Errorf("FAIL%s: unexpected error type '%s','%s' -> %d != %d",
			route, body, printPath(args...), errT, exp)
		return
	}
}

func TestRouteInvokeNoop(t *testing.T) {
	hNoop := func() {}
	rNoop := checkRoute(t, hNoop, "noop", f("noop"))

	checkInvoke(t, rNoop, "", "")
	checkInvoke(t, rNoop, "", "123", p(0, "321"))
}

func TestRouteInvokeString(t *testing.T) {
	hStr := func(s string) string { return s + "a" }

	rStrBody := checkRoute(t, hStr, "str/body", f("str"), f("body"))
	checkInvoke(t, rStrBody, `"cba"`, `"cb"`)
	failInvoke(t, rStrBody, UnmarshalError, ``)

	rStrArg := checkRoute(t, hStr, "str/{0:arg}", f("str"), p(0, "arg"))
	checkInvoke(t, rStrArg, `"cba"`, ``, p(0, "cb"))
	checkInvoke(t, rStrArg, `"cba"`, `"blah"`, p(0, "cb"))
}

func TestRouteInvokeBool(t *testing.T) {
	hBool := func(b bool) bool { return !b }

	rBoolBody := checkRoute(t, hBool, "bool/body", f("bool"), f("body"))
	checkInvoke(t, rBoolBody, "false", "true")
	failInvoke(t, rBoolBody, UnmarshalError, "")
	failInvoke(t, rBoolBody, UnmarshalError, "abc")

	rBoolArg := checkRoute(t, hBool, "bool/{0:arg}", f("bool"), p(0, "arg"))
	checkInvoke(t, rBoolArg, "false", "", p(0, "true"))
	checkInvoke(t, rBoolArg, "false", "false", p(0, "true"))
	failInvoke(t, rBoolArg, UnmarshalError, "", p(0, "abc"))
}

func TestRouteInvokeInt(t *testing.T) {
	hInt := func(i int) int { return i + 1 }

	rIntBody := checkRoute(t, hInt, "int/body", f("int"), f("body"))
	checkInvoke(t, rIntBody, "124", "123")
	failInvoke(t, rIntBody, UnmarshalError, "")
	failInvoke(t, rIntBody, UnmarshalError, "abc")
	failInvoke(t, rIntBody, UnmarshalError, "1.2")

	rIntArg := checkRoute(t, hInt, "int/{0:arg}", f("int"), p(0, "arg"))
	checkInvoke(t, rIntArg, "124", "", p(0, "123"))
	checkInvoke(t, rIntArg, "124", "321", p(0, "123"))
	failInvoke(t, rIntArg, UnmarshalError, "", p(0, "abc"))
	failInvoke(t, rIntArg, UnmarshalError, "", p(0, "1.2"))
}

func TestRouteInvokeUint(t *testing.T) {
	hUint := func(u uint) uint { return u + 1 }

	rUintBody := checkRoute(t, hUint, "uint/body", f("uint"), f("body"))
	checkInvoke(t, rUintBody, "124", "123")
	failInvoke(t, rUintBody, UnmarshalError, "")
	failInvoke(t, rUintBody, UnmarshalError, "abc")
	failInvoke(t, rUintBody, UnmarshalError, "-123")

	rUintArg := checkRoute(t, hUint, "uint/{0:arg}", f("uint"), p(0, "arg"))
	checkInvoke(t, rUintArg, "124", "", p(0, "123"))
	checkInvoke(t, rUintArg, "124", "321", p(0, "123"))
	failInvoke(t, rUintArg, UnmarshalError, "", p(0, "abc"))
	failInvoke(t, rUintArg, UnmarshalError, "", p(0, "-123"))
}

func TestRouteInvokeFloat(t *testing.T) {
	hFloat := func(f float64) float64 { return f + 1 }

	rFloatBody := checkRoute(t, hFloat, "float/body", f("float"), f("body"))
	checkInvoke(t, rFloatBody, "124.5", "123.5")
	failInvoke(t, rFloatBody, UnmarshalError, "")
	failInvoke(t, rFloatBody, UnmarshalError, "abc")

	rFloatArg := checkRoute(t, hFloat, "float/{0:arg}", f("float"), p(0, "arg"))
	checkInvoke(t, rFloatArg, "124.5", "", p(0, "123.5"))
	checkInvoke(t, rFloatArg, "124.5", "321.1", p(0, "123.5"))
	failInvoke(t, rFloatArg, UnmarshalError, "", p(0, "abc"))
}

func TestRouteInvokeObj(t *testing.T) {
	hObj := func(t T) T { return T{t.Value + 1} }

	rObj := checkRoute(t, hObj, "obj/body", f("obj"), f("body"))
	checkInvoke(t, rObj, `{"val":124}`, `{"val":123}`)
	checkInvoke(t, rObj, `{"val":1}`, `{}`)
	failInvoke(t, rObj, UnmarshalError, ``)
	failInvoke(t, rObj, UnmarshalError, `{"val":123`)
}

func TestRouteInvokePtr(t *testing.T) {
	hPtr := func(t *T) *T { return &T{t.Value + 1} }

	rPtr := checkRoute(t, hPtr, "ptr/body", f("ptr"), f("body"))
	checkInvoke(t, rPtr, `{"val":124}`, `{"val":123}`)
	checkInvoke(t, rPtr, `{"val":1}`, `{}`)
	failInvoke(t, rPtr, UnmarshalError, ``)
	failInvoke(t, rPtr, UnmarshalError, `{"val":123`)
}

func TestRouteInvokeMulti(t *testing.T) {
	hMulti := func(a, b, c int) int { return a + b + c }

	rMulti0 := checkRoute(t, hMulti, "multi/{0:a}/{1:b}/{2:c}", f("multi"), p(0, "a"), p(1, "b"), p(2, "c"))
	checkInvoke(t, rMulti0, "123", "", p(0, "100"), p(1, "20"), p(2, "3"))
	checkInvoke(t, rMulti0, "123", "", p(2, "3"), p(1, "20"), p(0, "100"))
	failInvoke(t, rMulti0, UnmarshalError, "", p(0, "100"), p(1, "20"), p(2, "a"))
	failInvoke(t, rMulti0, UnmarshalError, "", p(2, "a"), p(1, "20"), p(0, "100"))

	rMulti1 := checkRoute(t, hMulti, "multi/{0:a}/{2:c}", f("multi"), p(0, "a"), p(2, "c"))
	checkInvoke(t, rMulti1, "123", "20", p(0, "100"), p(2, "3"))
	checkInvoke(t, rMulti1, "123", "20", p(2, "3"), p(0, "100"))
	failInvoke(t, rMulti1, UnmarshalError, "", p(2, "a"), p(0, "100"))
	failInvoke(t, rMulti1, UnmarshalError, "a", p(2, "a"), p(0, "100"))
	failInvoke(t, rMulti1, UnmarshalError, "20", p(0, "100"), p(2, "a"))
	failInvoke(t, rMulti1, UnmarshalError, "20", p(2, "a"), p(0, "100"))
}

func TestRouteInvokeError(t *testing.T) {
	hErr0 := func() error { return fmt.Errorf("BOOM") }
	rErr0 := checkRoute(t, hErr0, "err/0", f("err"), f("0"))
	failInvoke(t, rErr0, HandlerError, "")

	hErr1 := func() (int, error) { return 0, fmt.Errorf("BOOM") }
	rErr1 := checkRoute(t, hErr1, "err/1", f("err"), f("1"))
	failInvoke(t, rErr1, HandlerError, "")

	hErr2 := func() (error, int) { return fmt.Errorf("BOOM"), 0 }
	rErr2 := checkRoute(t, hErr2, "err/2", f("err"), f("2"))
	failInvoke(t, rErr2, HandlerError, "")
}
