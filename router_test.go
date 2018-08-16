package iafon

import (
	"net/http"
	"strings"
	"testing"
)

func TestServeHTTP(t *testing.T) {
	var echo string

	r := newRouter()

	method := "GET"
	path := "/func"

	r.Handle(method, path, func(c *Context) {
		echo = c.Req.URL.Path
	})

	req, _ := http.NewRequest(method, "http://localhost"+path, nil)

	r.ServeHTTP(nil, req)

	if echo != path {
		t.Fatal("func(*Context) is not fired on request")
	}
}

type MockResponseWriter struct {
	header http.Header
	code   int
	data   string
}

func (rsp *MockResponseWriter) Header() http.Header {
	return rsp.header
}

func (rsp *MockResponseWriter) Write(data []byte) (int, error) {
	rsp.data = strings.TrimSpace(string(data))
	return len(rsp.data), nil
}

func (rsp *MockResponseWriter) WriteHeader(statusCode int) {
	rsp.code = statusCode
}

func TestRedirect(t *testing.T) {
	var handler = func(c *Context) {}

	r := newRouter()

	r.GET("/user/", handler)
	r.GET("/user/:name/", handler)
	r.GET("/user/admin/", handler)
	r.GET("host/user/", handler)
	r.GET("host/user/:name/", handler)
	r.GET("host/user/admin/", handler)

	var paths = map[string]string{
		"localhost/user":       `<a href="http://localhost/user/">Temporary Redirect</a>.`,
		"localhost/user/lwj":   `<a href="http://localhost/user/lwj/">Temporary Redirect</a>.`,
		"localhost/user/admin": `<a href="http://localhost/user/admin/">Temporary Redirect</a>.`,
		"host/user":            `<a href="http://host/user/">Temporary Redirect</a>.`,
		"host/user/lwj":        `<a href="http://host/user/lwj/">Temporary Redirect</a>.`,
		"host/user/admin":      `<a href="http://host/user/admin/">Temporary Redirect</a>.`,
	}

	for path, redirect_to := range paths {
		rsp := &MockResponseWriter{header: http.Header{}}
		req, _ := http.NewRequest("GET", "http://"+path, nil)
		r.ServeHTTP(rsp, req)

		if rsp.code != http.StatusTemporaryRedirect || rsp.data != redirect_to {
			t.Fatalf("fail to redirect. %s should redirect to %s\ngot:%s\n", path, redirect_to, rsp.data)
		}
	}
}
