package iafon

import (
	"net/http"
	"strings"
	"testing"
)

func TestSetPrefix(t *testing.T) {
	r := newRouter()

	if r.prefix != "" {
		t.Fatal("newRouter has non empty prefix")
	}

	{
		r := r.Group()
		r.SetPrefix("/user")
		g := r.Group()
		g.SetPrefix("/profile")
		sg := g.Group()
		sg.SetPrefix("/iafon")
		sg1 := g.Group()
		sg1.SetPrefix("/iafon1")

		if r.prefix != "/user" {
			t.Fatal("SetPrefix fail to modify prefix")
		}

		if g.prefix != "/user/profile" {
			t.Fatal("SetPrefix fail to concat two prefix")
		}

		if sg.prefix != "/user/profile/iafon" {
			t.Fatal("SetPrefix fail to concat three prefix")
		}

		if sg1.prefix != "/user/profile/iafon1" {
			t.Fatal("SetPrefix fail to concat three prefix 2")
		}
	}

	{
		r := r.Group()
		r.SetPrefix("user")
		g := r.Group()
		g.SetPrefix("/profile")
		sg := g.Group()
		sg.SetPrefix("/iafon")
		sg1 := g.Group()
		sg1.SetPrefix("/iafon1")

		if r.prefix != "user" {
			t.Fatal("SetPrefix fail to modify prefix")
		}

		if g.prefix != "user/profile" {
			t.Fatal("SetPrefix fail to concat two prefix")
		}

		if sg.prefix != "user/profile/iafon" {
			t.Fatal("SetPrefix fail to concat three prefix")
		}

		if sg1.prefix != "user/profile/iafon1" {
			t.Fatal("SetPrefix fail to concat three prefix 2")
		}
	}
}

func TestSetPrefixByGroupMethod(t *testing.T) {
	r := newRouter()

	if r.prefix != "" {
		t.Fatal("newRouter has non empty prefix")
	}

	r.SetPrefix("/user")
	g := r.Group("/profile")
	sg := g.Group("/iafon")
	sg1 := g.Group("/iafon1")

	if r.prefix != "/user" {
		t.Fatal("SetPrefix fail to modify prefix")
	}

	if g.prefix != "/user/profile" {
		t.Fatal("SetPrefix fail to concat two prefix")
	}

	if sg.prefix != "/user/profile/iafon" {
		t.Fatal("SetPrefix fail to concat three prefix")
	}

	if sg1.prefix != "/user/profile/iafon1" {
		t.Fatal("SetPrefix fail to concat three prefix 2")
	}
}

func TestPrefixConcatPattern(t *testing.T) {
	var rn *RouteNode

	r := newRouter()
	r.SetPrefix("/user")

	rn = r.Handle("GET", "/1", func(*Context) {})
	if rn.pattern != "/user/1" {
		t.Fatal("prefix '/user' concat pattern '/1' fail to be '/user/1'")
	}

	g := r.Group()
	g.SetPrefix("/profile")

	rn = g.Handle("GET", "/1", func(*Context) {})
	if rn.pattern != "/user/profile/1" {
		t.Fatal("prefix '/user' and '/profile' concat pattern '/1' fail to be '/user/profile/1'")
	}
}

var execSequence []int
var stopMiddlewareIndex int

type TestMassMiddleware struct {
	Middleware
	index int
}

func (m *TestMassMiddleware) Handle() bool {
	execSequence = append(execSequence, m.index)
	if stopMiddlewareIndex > 0 && m.index == stopMiddlewareIndex {
		return false
	} else {
		return true
	}
}

func newMiddleware(index int, execOrder ...int16) *TestMassMiddleware {
	var order int16 = 0
	if len(execOrder) > 0 {
		order = execOrder[0]
	}
	return &TestMassMiddleware{index: index, Middleware: Middleware{ExecOrder: order}}
}

func handleFunc(c *Context) {
	execSequence = append(execSequence, 0)
}

func TestTwoLevelGroupForm1(t *testing.T) {
	method := "GET"
	path := "/group/group/route"

	r := newRouter()

	r.UseMiddleware(newMiddleware(1))
	r.UseMiddleware(newMiddleware(2))

	r.Group("/group",
		newMiddleware(11),
		newMiddleware(12, 1),
		func(r *RouteGroup) {
			r.Group("/group",
				newMiddleware(31, -1),
				newMiddleware(32, -1),
				func(r *RouteGroup) {
					rn := r.Handle(method, "/route", handleFunc)
					rn.UseMiddleware(newMiddleware(41), -2)
					rn.UseMiddleware(newMiddleware(42), -1)
				},
			)
		},
	)

	correctSequence := []int{12, 1, 2, 11, 0, 31, 32, 42, 41}
	execSequence = []int{}

	req, _ := http.NewRequest(method, "http://localhost"+path, nil)
	r.ServeHTTP(nil, req)

	if len(execSequence) != len(correctSequence) {
		t.Fatalf("some middlewares are not executed. expected %d, got %d.", len(correctSequence), len(execSequence))
	}

	for i := range execSequence {
		if execSequence[i] != correctSequence[i] {
			t.Fatalf("middlewares are executed in incorrect order. %v", execSequence)
		}
	}
}

func TestTwoLevelGroupForm2(t *testing.T) {
	method := "GET"
	path := "/group/group/route"

	r := newRouter()
	r.UseMiddleware(newMiddleware(1))
	r.UseMiddleware(newMiddleware(2))
	{
		r := r.Group("/group")
		r.UseMiddleware(newMiddleware(11))
		r.UseMiddleware(newMiddleware(12), 1)
		{
			r := r.Group("/group")
			r.UseMiddleware(newMiddleware(31), -1)
			r.UseMiddleware(newMiddleware(32), -1)

			rn := r.Handle(method, "/route", handleFunc)
			rn.UseMiddleware(newMiddleware(41), -2)
			rn.UseMiddleware(newMiddleware(42), -1)
		}
	}

	correctSequence := []int{12, 1, 2, 11, 0, 31, 32, 42, 41}
	execSequence = []int{}

	req, _ := http.NewRequest(method, "http://localhost"+path, nil)
	r.ServeHTTP(nil, req)

	if len(execSequence) != len(correctSequence) {
		t.Fatalf("some middlewares are not executed. expected %d, got %d.", len(correctSequence), len(execSequence))
	}

	for i := range execSequence {
		if execSequence[i] != correctSequence[i] {
			t.Fatalf("middlewares are executed in incorrect order. %v", execSequence)
		}
	}
}

func TestStopMiddleware(t *testing.T) {
	method := "GET"
	path := "/group/group/route"

	r := newRouter()
	r.UseMiddleware(newMiddleware(1))
	r.UseMiddleware(newMiddleware(2))
	{
		r := r.Group("/group")
		r.UseMiddleware(newMiddleware(11))
		r.UseMiddleware(newMiddleware(12), 1)
		{
			r := r.Group("/group")
			r.UseMiddleware(newMiddleware(31), -1)
			r.UseMiddleware(newMiddleware(32), -1)

			rn := r.Handle(method, "/route", handleFunc)
			rn.UseMiddleware(newMiddleware(41, -2))
			rn.UseMiddleware(newMiddleware(42, -1))
		}
	}

	stopMiddlewareIndex = 2
	correctSequence := []int{12, 1, 2}
	execSequence = []int{}

	req, _ := http.NewRequest(method, "http://localhost"+path, nil)
	r.ServeHTTP(nil, req)

	if len(execSequence) != len(correctSequence) {
		t.Fatalf("some middlewares are not executed. expected %d, got %d.", len(correctSequence), len(execSequence))
	}

	for i := range execSequence {
		if execSequence[i] != correctSequence[i] {
			t.Fatalf("middlewares are executed in incorrect order. %v", execSequence)
		}
	}
}

func TestStopMiddleware2(t *testing.T) {
	method := "GET"
	path := "/group/group/route"

	r := newRouter()
	r.UseMiddleware(newMiddleware(1))
	r.UseMiddleware(newMiddleware(2))
	{
		r := r.Group("/group")
		r.UseMiddleware(newMiddleware(11))
		r.UseMiddleware(newMiddleware(12, 1))
		{
			r := r.Group("/group")
			r.UseMiddleware(newMiddleware(31), -1)
			r.UseMiddleware(newMiddleware(32), -1)

			rn := r.Handle(method, "/route", handleFunc)
			rn.UseMiddleware(newMiddleware(41, -2))
			rn.UseMiddleware(newMiddleware(42, -1))
		}
	}

	stopMiddlewareIndex = 31
	correctSequence := []int{12, 1, 2, 11, 0, 31}
	execSequence = []int{}

	req, _ := http.NewRequest(method, "http://localhost"+path, nil)
	r.ServeHTTP(nil, req)

	if len(execSequence) != len(correctSequence) {
		t.Fatalf("some middlewares are not executed. expected %d, got %d.", len(correctSequence), len(execSequence))
	}

	for i := range execSequence {
		if execSequence[i] != correctSequence[i] {
			t.Fatalf("middlewares are executed in incorrect order. %v", execSequence)
		}
	}
}

func TestHttpMethods(t *testing.T) {
	var echoMethod string
	var handlerFactory = func(method string) func(*Context) {
		return func(*Context) {
			echoMethod = method
		}
	}

	r := newRouter()
	r.GET("/", handlerFactory("GET"))
	r.POST("/", handlerFactory("POST"))
	r.PUT("/", handlerFactory("PUT"))
	r.DELETE("/", handlerFactory("DELETE"))
	r.OPTIONS("/", handlerFactory("OPTIONS"))
	r.HEAD("/", handlerFactory("HEAD"))
	r.PATCH("/", handlerFactory("PATCH"))
	r.Handle("CONNECT", "/", handlerFactory("CONNECT"))
	r.Handle("TRACE", "/", handlerFactory("TRACE"))

	for m := range http_methods {
		if m != "*" {
			req, _ := http.NewRequest(m, "http://localhost/", nil)
			r.ServeHTTP(nil, req)

			if echoMethod != m {
				t.Fatalf("%s fail to be handled.", m)
			}
		}
	}
}

func TestAnyHttpMethod(t *testing.T) {
	var echoMethod string
	var handlerFactory = func(method string) func(*Context) {
		return func(*Context) {
			echoMethod = method
		}
	}

	r := newRouter()
	r.GET("/", handlerFactory("GET"))
	r.Any("/", handlerFactory("*"))

	for m := range http_methods {
		if m != "*" {
			req, _ := http.NewRequest(m, "http://localhost/", nil)
			r.ServeHTTP(nil, req)

			if (m == "GET" && echoMethod != "GET") || (m != "GET" && echoMethod != "*") {
				t.Fatalf("%s fail to be handled.", m)
			}
		}
	}
}

func TestSomeHttpMethod(t *testing.T) {
	var echoMethod string
	var handlerFactory = func(method string) func(*Context) {
		return func(*Context) {
			echoMethod = method
		}
	}

	r := newRouter()
	r.Some([]string{"GET", "POST"}, "/", handlerFactory("GET_POST"))
	r.Any("/", handlerFactory("*"))

	for m := range http_methods {
		if m != "*" {
			req, _ := http.NewRequest(m, "http://localhost/", nil)
			r.ServeHTTP(nil, req)

			if ((m == "GET" || m == "POST") && echoMethod != "GET_POST") ||
				(m != "GET" && m != "POST" && echoMethod != "*") {
				t.Fatalf("%s fail to be handled.", m)
			}
		}
	}
}

func TestDuplicateRoute(t *testing.T) {
	defer func() {
		if p := recover(); p != nil {
			if p != "http: duplicate route 'GET /test'" {
				t.Fatal("add duplicate route should panic")
			}
		}
	}()

	r := newRouter()

	r.GET("/test", func(*Context) {})
	r.GET("/test", func(*Context) {})
}

func TestRoutePriority(t *testing.T) {
	var echo string

	var handlerUserIndex = func(c *Context) {
		echo = "/user"
	}
	var handlerUserAdmin = func(c *Context) {
		echo = "/user/admin"
	}
	var handlerUserAdministrator = func(c *Context) {
		echo = "/user/administrator"
	}
	var handlerUser = func(c *Context) {
		echo = "/user/:name"
	}
	var handlerUser1 = func(c *Context) {
		echo = "/user/:name/"
	}
	var handlerHostUserIndex = func(c *Context) {
		echo = "host/user"
	}
	var handlerHostUserAdmin = func(c *Context) {
		echo = "host/user/admin/"
	}
	var handlerHostUser = func(c *Context) {
		echo = "host/user/:name"
	}
	var handlerHostUser1 = func(c *Context) {
		echo = "host/user/:name/"
	}

	r := newRouter()
	r.HandleError(404, func(*Context) {})

	r.GET("/user", handlerUserIndex)
	r.GET("/user/admin", handlerUserAdmin)
	r.GET("/user/:name", handlerUser)
	r.GET("/user/:name/", handlerUser1)
	r.GET("/user/administrator", handlerUserAdministrator)
	r.GET("host/user", handlerHostUserIndex)
	r.GET("host/user/admin/", handlerHostUserAdmin)
	r.GET("host/user/:name", handlerHostUser)
	r.GET("host/user/:name/", handlerHostUser1)

	var paths = map[string]string{
		"/user":               "/user",
		"/user/":              "/user",
		"/user/admin":         "/user/admin",
		"/user/admin/":        "/user/admin",
		"/user/administrator": "/user/administrator",
		"/user/adm":           "/user/:name",
		"/user/administ":      "/user/:name",
		"/user/test":          "/user/:name",
		"/user/test/":         "/user/:name/",
		"/user/adm/":          "/user/:name/",
	}

	for path, pattern := range paths {
		echo = ""

		req, _ := http.NewRequest("GET", "http://localhost"+path, nil)
		r.ServeHTTP(nil, req)

		if path == "/user/admin/" {
			if echo != "" {
				t.Fatalf("route should not be executed. req path: %s, echo pattern: %s", path, echo)
			}
		} else if echo != pattern {
			t.Fatalf("route is not executed in priority. req path: %s, echo pattern: %s", path, echo)
		}
	}

	paths = map[string]string{
		"host/user":               "host/user",
		"host/user/":              "host/user",
		"host/user/admin":         "/user/admin",
		"host/user/admin/":        "host/user/admin/",
		"host/user/administrator": "/user/administrator",
		"host/user/adm":           "host/user/:name",
		"host/user/administ":      "host/user/:name",
		"host/user/test":          "host/user/:name",
		"host/user/adm/":          "host/user/:name/",
	}

	for path, pattern := range paths {
		echo = ""

		req, _ := http.NewRequest("GET", "http://"+path, nil)
		r.ServeHTTP(nil, req)

		if echo != pattern {
			t.Fatalf("route with host is not executed in priority. req path: %s, echo pattern: %s", path, echo)
		}
	}
}

func TestGetRoutes(t *testing.T) {
	var handler = func(*Context) {}

	var routes = map[string]bool{
		"/user POST":             true,
		"/user GET":              true,
		"host/user/:name PUT":    true,
		"/user/admin GET":        true,
		"/user/:name PUT":        true,
		"host/user/admin PUT":    true,
		"/user/admin PUT":        true,
		"/user/:name GET":        true,
		"/user/:name DELETE":     true,
		"host/user POST":         true,
		"host/user/:name DELETE": true,
	}

	r := newRouter()

	for v := range routes {
		parts := strings.Split(v, " ")
		r.Handle(parts[1], parts[0], handler)
	}

	get_routes := r.GetRoutes()

	if len(routes) != len(get_routes) {
		t.Fatal("GetRoutes error, count error")
	}

	for _, v := range get_routes {
		if routes[v.host+v.pattern+" "+v.method] == false {
			t.Fatal("GetRoutes error")
		}
	}

	routesString :=
		`GET         /user
POST        /user
DELETE      /user/:name
GET         /user/:name
PUT         /user/:name
GET         /user/admin
PUT         /user/admin
POST   host /user
DELETE host /user/:name
PUT    host /user/:name
PUT    host /user/admin
`

	if strings.TrimSpace(get_routes.String()) != strings.TrimSpace(routesString) {
		t.Fatal("GetRoutes().Stirng() error")
	}
}
