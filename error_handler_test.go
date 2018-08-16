package iafon

import (
	"net/http"
	"testing"
)

var error_handler_test_echo string

type TestErrorHandler404 struct{}

func (h *TestErrorHandler404) Handle(c *Context) {
	error_handler_test_echo = "404"
}

type TestErrorHandler405 struct{}

func (h *TestErrorHandler405) Handle(c *Context) {
	error_handler_test_echo = "405"
}

type TestErrorHandler500 struct{}

func (h *TestErrorHandler500) Handle(c *Context) {
	error_handler_test_echo = "500"
}

func TestHandler404(t *testing.T) {
	r := newRouter()
	r.HandleError(404, &TestErrorHandler404{})
	r.GET("/test", func(*Context) {})

	error_handler_test_echo = ""

	req, _ := http.NewRequest("GET", "http://localhost/test404", nil)
	r.ServeHTTP(nil, req)

	if error_handler_test_echo != "404" {
		t.Fatal("404 Handler is not fired on request")
	}
}

func TestHandler405(t *testing.T) {
	r := newRouter()
	r.HandleError(405, &TestErrorHandler405{})
	r.GET("/test", func(*Context) {})

	error_handler_test_echo = ""

	req, _ := http.NewRequest("POST", "http://localhost/test", nil)
	r.ServeHTTP(nil, req)

	if error_handler_test_echo != "405" {
		t.Fatal("405 Handler is not fired on request")
	}
}

func TestHandler500(t *testing.T) {
	r := newRouter()
	r.HandleError(500, &TestErrorHandler500{})
	r.GET("/test", func(*Context) {
		panic("trigger 500")
	})

	error_handler_test_echo = ""

	req, _ := http.NewRequest("GET", "http://localhost/test", nil)
	r.ServeHTTP(nil, req)

	if error_handler_test_echo != "500" {
		t.Fatal("500 Handler is not fired on request")
	}
}
