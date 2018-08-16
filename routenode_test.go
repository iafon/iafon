package iafon

import (
	"net/http"
	"testing"
)

var routenode_test_echo string

type TestRouteNodeMiddleware struct {
	Middleware
}

func (t *TestRouteNodeMiddleware) Handle() bool {
	routenode_test_echo = t.Req.URL.Path
	return true
}

func TestUseMiddlewareOnRouteNode(t *testing.T) {
	s := NewServer("127.0.0.1:")

	method := "GET"
	path := "/routenode/middleware"

	rn := s.Handle(method, path, func(*Context) {})
	rn.UseMiddleware(&TestRouteNodeMiddleware{})

	req, _ := http.NewRequest(method, "http://localhost"+path, nil)

	s.ServeHTTP(nil, req)

	if routenode_test_echo != path {
		t.Fatal("RouteNode Middleware is not fired on request")
	}
}
