package iafon

import (
	"net/http"
	"testing"
)

var exec_count = 0
var error_count = 0
var udata_error_count = 0

type TestContextMiddleware struct {
	Middleware
	index int
}

func (m *TestContextMiddleware) Handle() bool {
	if m.Param["groupname"] != "group1" || m.Param["name"] != "name1" {
		error_count++
	}
	exec_count++
	if m.Udata["pass"] != "test_udata" {
		udata_error_count++
	}
	if m.index == 2 {
		m.Udata = make(map[string]interface{})
		m.Udata["pass"] = "test_udata"
	}
	return true
}

func newCtxMiddleware(index int, execOrder ...int16) *TestContextMiddleware {
	var order int16 = 0
	if len(execOrder) > 0 {
		order = execOrder[0]
	}
	return &TestContextMiddleware{index: index, Middleware: Middleware{ExecOrder: order}}
}

func handleCtxFunc(c *Context) {
	if c.Param["groupname"] != "group1" || c.Param["name"] != "name1" {
		error_count++
	}
	if c.Udata["pass"] != "test_udata" {
		udata_error_count++
	}
	exec_count++
}

func TestContextParam(t *testing.T) {
	method := "GET"
	path := "/group/group1/name1"

	r := newRouter()

	r.UseMiddleware(newCtxMiddleware(1))
	r.UseMiddleware(newCtxMiddleware(2))

	r.Group("/group",
		newCtxMiddleware(11),
		newCtxMiddleware(12, 1),
		func(r *RouteGroup) {
			r.Group("/:groupname",
				newCtxMiddleware(31, -1),
				newCtxMiddleware(32, -1),
				func(r *RouteGroup) {
					rn := r.Handle(method, "/:name", handleCtxFunc)
					rn.UseMiddleware(newCtxMiddleware(41), -2)
					rn.UseMiddleware(newCtxMiddleware(42), -1)
				},
			)
		},
	)

	exec_count = 0
	error_count = 0

	req, _ := http.NewRequest(method, "http://localhost"+path, nil)
	r.ServeHTTP(nil, req)

	if exec_count != 9 || error_count > 0 {
		t.Fatalf("Context.Param error")
	}
}

func TestContextUdata(t *testing.T) {
	method := "GET"
	path := "/group/group1/name1"

	r := newRouter()

	r.UseMiddleware(newCtxMiddleware(1))
	r.UseMiddleware(newCtxMiddleware(2), 100)

	r.Group("/group",
		newCtxMiddleware(11),
		newCtxMiddleware(12, 1),
		func(r *RouteGroup) {
			r.Group("/:groupname",
				newCtxMiddleware(31, -1),
				newCtxMiddleware(32, -1),
				func(r *RouteGroup) {
					rn := r.Handle(method, "/:name", handleCtxFunc)
					rn.UseMiddleware(newCtxMiddleware(41), -2)
					rn.UseMiddleware(newCtxMiddleware(42), -1)
				},
			)
		},
	)

	exec_count = 0
	udata_error_count = 0

	req, _ := http.NewRequest(method, "http://localhost"+path, nil)
	r.ServeHTTP(nil, req)

	if exec_count != 9 || udata_error_count != 1 {
		t.Fatalf("Context.Udata error, exec_count: %d, udata_error_count: %d", exec_count, udata_error_count)
	}
}
