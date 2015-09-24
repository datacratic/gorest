// Copyright (c) 2014 Datacratic. All rights reserved.

package rest

import (
	"fmt"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type KV struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

type TestService struct {
	Map map[string]string

	mutex      sync.Mutex
	initialize sync.Once

	eventC chan int
}

func (service *TestService) Init() {
	service.initialize.Do(service.init)
}

func (service *TestService) init() {
	service.Map = make(map[string]string)
	service.eventC = make(chan int, 100)
}

func (service *TestService) RESTRoutes() Routes {
	return Routes{
		NewRoute("/map", "POST", service.Post),
		NewRoute("/map/:key", "GET", service.Get),
		NewRoute("/map/:key", "PUT", service.Put),
		NewRoute("/map/:key", "DELETE", service.Del),
	}
}

func (service *TestService) Get(key string) (*KV, error) {
	service.Init()
	service.mutex.Lock()
	defer service.mutex.Unlock()

	service.eventC <- 1

	if val, ok := service.Map[key]; ok {
		return &KV{key, val}, nil
	}

	return nil, fmt.Errorf("unknown key: %s", key)
}

func (service *TestService) Post(kv KV) (err error) {
	service.Init()
	service.mutex.Lock()
	defer service.mutex.Unlock()

	service.eventC <- 1

	if val, ok := service.Map[kv.Key]; ok {
		err = fmt.Errorf("existing: %s -> %s", kv.Key, val)
	} else {
		service.Map[kv.Key] = kv.Val
	}

	return
}

func (service *TestService) Put(key, val string) (err error) {
	service.Init()
	service.mutex.Lock()
	defer service.mutex.Unlock()

	service.eventC <- 1

	if _, ok := service.Map[key]; ok {
		service.Map[key] = val
	} else {
		err = fmt.Errorf("unknown key: %s", key)
	}

	return
}

func (service *TestService) Del(key string) (*KV, error) {
	service.Init()
	service.mutex.Lock()
	defer service.mutex.Unlock()

	service.eventC <- 1

	if val, ok := service.Map[key]; ok {
		delete(service.Map, key)
		return &KV{key, val}, nil
	}

	return nil, fmt.Errorf("unknown key: %s", key)
}

func (service *TestService) Wait(n int) int {
	service.Init()

	timeout := time.After(10 * time.Millisecond)

	for i := 0; i < n; {
		select {
		case <-timeout:
			return i
		case c := <-service.eventC:
			i += c
		}
	}

	return n
}

func (service *TestService) Expect(t *testing.T, title string, exp ...KV) {
	service.Init()
	service.mutex.Lock()
	defer service.mutex.Unlock()

	if len(service.Map) != len(exp) {
		t.Errorf("FAIL(%s): unexpected map length: %d != %d",
			title, len(service.Map), len(exp))
	}

	for _, kv := range exp {

		if val, ok := service.Map[kv.Key]; !ok {
			t.Errorf("FAIL(%s): missing key: %s", title, kv.Key)

		} else if val != kv.Val {
			t.Errorf("FAIL(%s): value mismatch for key %s: %s != %s", title, kv.Key, kv.Val, val)
		}
	}
}

func checkResp(t *testing.T, title string, resp *Response) {
	if err := resp.GetBody(nil); err != nil {
		t.Errorf("FAIL(%s): error %s", title, err)
	}
}

func checkRespBody(t *testing.T, title string, resp *Response, exp *KV) {
	var kv KV

	if err := resp.GetBody(&kv); err != nil {
		t.Errorf("FAIL(%s): error %s", title, err)

	} else if kv.Key != exp.Key || kv.Val != exp.Val {
		t.Errorf("FAIL(%s): value mismatch '%s:%s' != '%s:%s'",
			title, exp.Key, exp.Val, kv.Key, kv.Val)
	}
}

func failResp(t *testing.T, title string, resp *Response, exp ErrorType, code int) {
	if resp.Code != code {
		t.Errorf("FAIL(%s): unexpected code: %d != %d", title, resp.Code, code)
	}

	if err := resp.GetBody(nil); err.Type != exp {
		t.Errorf("FAIL(%s): unexpected error %s != %s:%v", title, exp, err.Type, err.Sub)
	}
}

func TestMuxSimple(t *testing.T) {
	handler := &TestService{}

	mux := new(Mux)
	mux.AddService(handler)

	fmt.Println(mux.GetRoutes())

	server := httptest.NewServer(mux)
	defer server.Close()

	client := &Client{Host: server.URL, Root: "/map/"}

	r00 := client.NewRequest("POST").SetBody(&KV{"a", "1"}).Send()
	checkResp(t, "p(a,1)", r00)

	r01 := client.NewRequest("POST").SetBody(&KV{"b", "2"}).Send()
	checkResp(t, "p(b,2)", r01)

	handler.Wait(2)
	handler.Expect(t, "r0x", KV{"a", "1"}, KV{"b", "2"})

	r10 := client.NewRequest("GET").SetPath("/%s", "a").Send()
	checkRespBody(t, "g(a)", r10, &KV{"a", "1"})

	r11 := client.NewRequest("PUT").SetPath("/%s", "b").SetBody("3").Send()
	checkResp(t, "u(b,3)", r11)

	handler.Wait(2)
	handler.Expect(t, "r1x", KV{"a", "1"}, KV{"b", "3"})

	r20 := client.NewRequest("POST").SetBody(&KV{"a", "4"}).Send()
	failResp(t, "p(a,4)", r20, EndpointError, 400)

	r21 := client.NewRequest("DELETE").SetPath("/b").Send()
	checkRespBody(t, "d(b)", r21, &KV{"b", "3"})

	r22 := client.NewRequest("POST").SetPath("/blah/bleh").Send()
	failResp(t, "p(blah)", r22, UnknownRoute, 404)

	handler.Wait(2)
	handler.Expect(t, "r2x", KV{"a", "1"})

}
