package iafon

import (
	"net/http"
	"testing"
)

var mix_handler_test_echo string

type TestHttpHandler struct{}

func (h *TestHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mix_handler_test_echo = r.URL.Path
}

func TestHandleHttpHandler(t *testing.T) {
	s := NewServer("127.0.0.1:")

	method := "GET"
	path := "/HttpHandler"

	rn := s.Handle(method, path, &TestHttpHandler{})
	if rn == nil {
		t.Fatal("fail to add http.Handler as handler")
	}

	req, _ := http.NewRequest(method, "http://localhost"+path, nil)

	s.ServeHTTP(nil, req)

	if mix_handler_test_echo != path {
		t.Fatal("http.Handler is not fired on request")
	}
}

func TestHandleHttpHandlerFunc(t *testing.T) {
	var echo string
	var f http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		echo = r.URL.Path
	}

	s := NewServer("127.0.0.1:")

	method := "GET"
	path := "/HandlerFunc"

	rn := s.Handle(method, path, f)
	if rn == nil {
		t.Fatal("fail to add http.HandlerFunc as handler")
	}

	req, _ := http.NewRequest(method, "http://localhost"+path, nil)

	s.ServeHTTP(nil, req)

	if echo != path {
		t.Fatal("http.HandlerFunc is not fired on request")
	}
}

func TestHandleRawHttpHandlerFunc(t *testing.T) {
	var echo string

	s := NewServer("127.0.0.1:")

	method := "GET"
	path := "/func"

	rn := s.Handle(method, path, func(w http.ResponseWriter, r *http.Request) {
		echo = r.URL.Path
	})
	if rn == nil {
		t.Fatal("fail to add func(http.ResponseWriter, *http.Request) as handler")
	}

	req, _ := http.NewRequest(method, "http://localhost"+path, nil)

	s.ServeHTTP(nil, req)

	if echo != path {
		t.Fatal("func(http.ResponseWriter, *http.Request) is not fired on request")
	}
}
