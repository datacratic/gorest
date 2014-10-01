// Copyright (c) 2014 Datacratic. All rights reserved.

package rest

import (
	"bytes"
	"fmt"
	"strconv"
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
	return PathItem{name, false}
}

func v(name string) PathItem {
	return PathItem{name, true}
}

func checkRoute(t *testing.T, handler interface{}, path string, exp ...PathItem) (route *Route) {
	route = NewRoute("POST", path, handler)

	if len(exp) != len(route.Path) {
		t.Errorf("FAIL: path length mismatch: %d:%s != %d:%s",
			len(exp), printPath(exp...), len(route.Path), printPath(route.Path...))
		return
	}

	for i, comp := range route.Path {
		if comp.Name != exp[i].Name || comp.IsArg != exp[i].IsArg {
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
	checkRoute(t, h1Arg, "/:a", v("a"))
	checkRoute(t, h1Arg, "a/:b/c", f("a"), v("b"), f("c"))

	failRoute(t, h1Arg, ":a/:b")

	h2Arg := func(T, int) (t T, err error) { return }

	failRoute(t, h2Arg, "")
	failRoute(t, h2Arg, "a")
	failRoute(t, h2Arg, "a/b")

	checkRoute(t, h2Arg, ":a", v("a"))
	checkRoute(t, h2Arg, ":a/:b", v("a"), v("b"))
	checkRoute(t, h2Arg, "c/:a/d/:b/e", f("c"), v("a"), f("d"), v("b"), f("e"))

	failRoute(t, h2Arg, ":a/:b/:c")

	failRoute(t, T{}, "")
	failRoute(t, &T{}, "")

	failRoute(t, func() (e0 error, e1 error) { return }, "")
	failRoute(t, func() (i0 int, i1 int) { return }, "")
	failRoute(t, func() (i0 int, i1 int, i2 int) { return }, "")
}

func checkInvoke(t *testing.T, route *Route, exp string, body string, args ...PathItem) {
	var m []string
	for _, arg := range args {
		m = append(m, arg.Name)
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
	var m []string
	for _, arg := range args {
		m = append(m, arg.Name)
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
	checkInvoke(t, rNoop, "", "123", v("321"))
}

func TestRouteInvokeString(t *testing.T) {
	hStr := func(s string) string { return s + "a" }

	rStrBody := checkRoute(t, hStr, "str/body", f("str"), f("body"))
	checkInvoke(t, rStrBody, `"cba"`, `"cb"`)
	failInvoke(t, rStrBody, UnmarshalError, ``)

	rStrArg := checkRoute(t, hStr, "str/:arg", f("str"), v("arg"))
	checkInvoke(t, rStrArg, `"cba"`, ``, v("cb"))
	checkInvoke(t, rStrArg, `"cba"`, `"blah"`, v("cb"))
}

func TestRouteInvokeBool(t *testing.T) {
	hBool := func(b bool) bool { return !b }

	rBoolBody := checkRoute(t, hBool, "bool/body", f("bool"), f("body"))
	checkInvoke(t, rBoolBody, "false", "true")
	failInvoke(t, rBoolBody, UnmarshalError, "")
	failInvoke(t, rBoolBody, UnmarshalError, "abc")

	rBoolArg := checkRoute(t, hBool, "bool/:arg", f("bool"), v("arg"))
	checkInvoke(t, rBoolArg, "false", "", v("true"))
	checkInvoke(t, rBoolArg, "false", "false", v("true"))
	failInvoke(t, rBoolArg, UnmarshalError, "", v("abc"))
}

func TestRouteInvokeInt(t *testing.T) {
	hInt := func(i int) int { return i + 1 }

	rIntBody := checkRoute(t, hInt, "int/body", f("int"), f("body"))
	checkInvoke(t, rIntBody, "124", "123")
	failInvoke(t, rIntBody, UnmarshalError, "")
	failInvoke(t, rIntBody, UnmarshalError, "abc")
	failInvoke(t, rIntBody, UnmarshalError, "1.2")

	rIntArg := checkRoute(t, hInt, "int/:arg", f("int"), v("arg"))
	checkInvoke(t, rIntArg, "124", "", v("123"))
	checkInvoke(t, rIntArg, "124", "321", v("123"))
	failInvoke(t, rIntArg, UnmarshalError, "", v("abc"))
	failInvoke(t, rIntArg, UnmarshalError, "", v("1.2"))
}

func TestRouteInvokeUint(t *testing.T) {
	hUint := func(u uint) uint { return u + 1 }

	rUintBody := checkRoute(t, hUint, "uint/body", f("uint"), f("body"))
	checkInvoke(t, rUintBody, "124", "123")
	failInvoke(t, rUintBody, UnmarshalError, "")
	failInvoke(t, rUintBody, UnmarshalError, "abc")
	failInvoke(t, rUintBody, UnmarshalError, "-123")

	rUintArg := checkRoute(t, hUint, "uint/:arg", f("uint"), v("arg"))
	checkInvoke(t, rUintArg, "124", "", v("123"))
	checkInvoke(t, rUintArg, "124", "321", v("123"))
	failInvoke(t, rUintArg, UnmarshalError, "", v("abc"))
	failInvoke(t, rUintArg, UnmarshalError, "", v("-123"))
}

func TestRouteInvokeFloat(t *testing.T) {
	hFloat := func(f float64) float64 { return f + 1 }

	rFloatBody := checkRoute(t, hFloat, "float/body", f("float"), f("body"))
	checkInvoke(t, rFloatBody, "124.5", "123.5")
	failInvoke(t, rFloatBody, UnmarshalError, "")
	failInvoke(t, rFloatBody, UnmarshalError, "abc")

	rFloatArg := checkRoute(t, hFloat, "float/:arg", f("float"), v("arg"))
	checkInvoke(t, rFloatArg, "124.5", "", v("123.5"))
	checkInvoke(t, rFloatArg, "124.5", "321.1", v("123.5"))
	failInvoke(t, rFloatArg, UnmarshalError, "", v("abc"))
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

	rMulti0 := checkRoute(t, hMulti, "multi/:a/:b/:c", f("multi"), v("a"), v("b"), v("c"))
	checkInvoke(t, rMulti0, "123", "", v("100"), v("20"), v("3"))
	checkInvoke(t, rMulti0, "123", "", v("3"), v("20"), v("100"))
	failInvoke(t, rMulti0, UnmarshalError, "", v("100"), v("20"), v("a"))
	failInvoke(t, rMulti0, UnmarshalError, "", v("a"), v("20"), v("100"))

	rMulti1 := checkRoute(t, hMulti, "multi/:a/:c", f("multi"), v("a"), v("c"))
	checkInvoke(t, rMulti1, "123", "20", v("100"), v("3"))
	checkInvoke(t, rMulti1, "123", "20", v("3"), v("100"))
	failInvoke(t, rMulti1, UnmarshalError, "", v("a"), v("100"))
	failInvoke(t, rMulti1, UnmarshalError, "a", v("a"), v("100"))
	failInvoke(t, rMulti1, UnmarshalError, "20", v("100"), v("a"))
	failInvoke(t, rMulti1, UnmarshalError, "20", v("a"), v("100"))
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

func BenchRouteInvoke(b *testing.B, route *Route, args []string, body []byte) {
	if _, _, err := route.invoke(args, body); err != nil {
		panic("failed bench")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		route.invoke(args, body)
	}
}

func BenchmarkRouteInvokeNoop(b *testing.B) {
	BenchRouteInvoke(b, NewRoute("POST", "", func() {}), nil, nil)
}

func BenchmarkRouteInvoke1Arg(b *testing.B) {
	args := []string{"10"}
	route := NewRoute("POST", ":a", func(int) {})

	BenchRouteInvoke(b, route, args, nil)
}

func BenchmarkRouteInvoke8Arg(b *testing.B) {
	path := "/"
	var args []string
	for i := 0; i < 8; i++ {
		args = append(args, strconv.Itoa(i))
		path += fmt.Sprintf(":%d/", i)
	}

	route := NewRoute("POST", path, func(a, b, c, d, e, f, g, h int) {})

	BenchRouteInvoke(b, route, args, nil)
}

func BenchmarkRouteInvokeBody(b *testing.B) {
	route := NewRoute("POST", "", func(a int) {})
	body := []byte("10")

	BenchRouteInvoke(b, route, nil, body)
}

func BenchmarkRouteInvokeRet(b *testing.B) {
	route := NewRoute("POST", "", func() int { return 10 })

	BenchRouteInvoke(b, route, nil, nil)
}

func BenchmarkRouteInvokeRetErr(b *testing.B) {
	route := NewRoute("POST", "", func() error { return nil })

	BenchRouteInvoke(b, route, nil, nil)
}

func BenchmarkRouteInvokeRetBoth(b *testing.B) {
	route := NewRoute("POST", "", func() (int, error) { return 10, nil })

	BenchRouteInvoke(b, route, nil, nil)
}

func BenchmarkRouteInvokeAll(b *testing.B) {
	path := "/"
	var args []string
	for i := 0; i < 7; i++ {
		args = append(args, strconv.Itoa(i))
		path += fmt.Sprintf(":%d/", i)
	}
	body := []byte("10")

	route := NewRoute("POST", path, func(a, b, c, d, e, f, g, h int) (int, error) { return 0, nil })

	BenchRouteInvoke(b, route, args, body)
}
