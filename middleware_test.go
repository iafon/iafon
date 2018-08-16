package iafon

import (
	"net/http"
	"testing"
)

var middleware_test_echo string

type TestMiddleware struct {
	Middleware
}

func (t *TestMiddleware) Handle() bool {
	middleware_test_echo = t.Req.URL.Path
	return true
}

var middleware_test_echo_1 string

type TestMiddleware1 struct {
	Middleware
}

func (t *TestMiddleware1) Handle() bool {
	middleware_test_echo_1 = t.Req.URL.Path
	return true
}

func TestUseMiddlewareGlobally(t *testing.T) {
	s := NewServer("127.0.0.1:")

	method := "GET"
	path := "/global/middleware"

	s.UseMiddleware(&TestMiddleware{})
	s.Handle(method, path, func(*Context) {})

	req, _ := http.NewRequest(method, "http://localhost"+path, nil)

	s.ServeHTTP(nil, req)

	if middleware_test_echo != path {
		t.Fatal("Global Middleware is not fired on request")
	}
}

func TestUseMiddlewareOnGroup(t *testing.T) {
	s := NewServer("127.0.0.1:")

	method := "GET"
	path := "/group/middleware"

	s.Handle(method, "/", func(*Context) {})
	s.Group("/group", &TestMiddleware{}, func(s *RouteGroup) {
		s.Handle(method, "/middleware", func(*Context) {})
	})

	req, _ := http.NewRequest(method, "http://localhost"+path, nil)

	s.ServeHTTP(nil, req)

	if middleware_test_echo != path {
		t.Fatal("Group Middleware is not fired on request")
	}
}

func TestUseMiddlewareOnRoute(t *testing.T) {
	s := NewServer("127.0.0.1:")

	method := "GET"
	path := "/route/middleware"

	s.Handle(method, path, func(*Context) {}).UseMiddleware(&TestMiddleware{})

	req, _ := http.NewRequest(method, "http://localhost"+path, nil)

	s.ServeHTTP(nil, req)

	if middleware_test_echo != path {
		t.Fatal("Route Middleware is not fired on request")
	}
}

func TestUseMultipleMiddleware(t *testing.T) {
	s := NewServer("127.0.0.1:")

	method := "GET"
	path := "/group/middleware/route"

	s.Handle(method, "/", func(*Context) {})
	s.Group("/group", &TestMiddleware{}, func(s *RouteGroup) {
		s.Handle(method, "/middleware/route", func(*Context) {}).UseMiddleware(&TestMiddleware1{})
		s.Handle(method, "/middleware/", func(*Context) {})
	})

	req, _ := http.NewRequest(method, "http://localhost"+path, nil)

	s.ServeHTTP(nil, req)

	if middleware_test_echo != path ||
		middleware_test_echo_1 != path {
		t.Fatal("Middleware is not fired on request, when using multi middleware")
	}
}

func TestUseMiddlewareAfterRouteAdded(t *testing.T) {
	s := NewServer("127.0.0.1:")

	method := "GET"
	path := "/group/middleware/after_route"

	s.Handle(method, "/", func(*Context) {})
	g := s.Group("/group", func(s *RouteGroup) {
		s.Handle(method, "/middleware", func(*Context) {})
	})
	g.UseMiddleware(&TestMiddleware{})

	req, _ := http.NewRequest(method, "http://localhost"+path, nil)

	s.ServeHTTP(nil, req)

	if middleware_test_echo != path {
		t.Fatal("Middleware is not fired on request, when adding after route added")
	}
}

func TestUseMiddlewareAsHandler(t *testing.T) {
	defer func() {
		if p := recover(); p == nil {
			t.Fatal("use middleware as route main handler should panic.")
		}
	}()

	r := newRouter()
	r.Handle("GET", "/", &TestMiddleware{})
}
