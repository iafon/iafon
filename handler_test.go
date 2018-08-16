package iafon

import (
	"net/http"
	"testing"
)

var handler_test_echo string

type TestHandler struct{}

func (h *TestHandler) Handle(c *Context) {
	handler_test_echo = c.Req.URL.Path
}

func TestHandleHandler(t *testing.T) {
	s := NewServer("127.0.0.1:")

	method := "GET"
	path := "/Handler"

	rn := s.Handle(method, path, &TestHandler{})
	if rn == nil {
		t.Fatal("fail to add Handler as handler")
	}

	req, _ := http.NewRequest(method, "http://localhost"+path, nil)

	s.ServeHTTP(nil, req)

	if handler_test_echo != path {
		t.Fatal("Handler is not fired on request")
	}
}

func TestHandleHandlerFunc(t *testing.T) {
	var echo string
	var f HandlerFunc = func(c *Context) {
		echo = c.Req.URL.Path
	}

	s := NewServer("127.0.0.1:")

	method := "GET"
	path := "/HandlerFunc"

	rn := s.Handle(method, path, f)
	if rn == nil {
		t.Fatal("fail to add HandlerFunc as handler")
	}

	req, _ := http.NewRequest(method, "http://localhost"+path, nil)

	s.ServeHTTP(nil, req)

	if echo != path {
		t.Fatal("HandlerFunc is not fired on request")
	}
}

func TestHandleRawHandlerFunc(t *testing.T) {
	var echo string

	s := NewServer("127.0.0.1:")

	method := "GET"
	path := "/func"

	rn := s.Handle(method, path, func(c *Context) {
		echo = c.Req.URL.Path
	})
	if rn == nil {
		t.Fatal("fail to add func(*Context) as handler")
	}

	req, _ := http.NewRequest(method, "http://localhost"+path, nil)

	s.ServeHTTP(nil, req)

	if echo != path {
		t.Fatal("func(*Context) is not fired on request")
	}
}
