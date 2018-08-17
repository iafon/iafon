# iafon ['jɑ:fon]
iafon's a framework or not.  
Basically, iafon is a http router written in go. But it's not only a http router, it support middlewares, controllers, route groups, route parameters, custom http error handlers. It's a lightweight http framework, not a web framework yet.

the following example will illuminate all iafon's features.

# 中文名
哑蜂

# example 

```
package main

import (
    "github.com/iafon/iafon"
    "net/http"
    "fmt"
)

func main() {
    s := iafon.NewServer(":8090")

    // we can set prefix on server, which will be applied to all routes
    // but usually we do not set prefix to all routes
    // s.SetPrefix("/api")

    // zero or more middleware could be used
    // the second parameter is execution order of middleware, default is 0
    // if execution order is negative, the middleware will be executed after route main handler
    // middlewares is executed according to order from high to low
    // for the same execution order, middlewares will be executed according to adding order
    //
    // if a Middleware return false, the middlewares and main handler after this middleware will not be executed
    // 
    // these two middlewares is used to all routes, because server is the top group contain all routes
    s.UseMiddleware(&AMiddleware{})
    s.UseMiddleware(&BMiddleware{}, 100)

    // main handler as iafon.Handler
    s.GET("/handler", &IafonHandler{})

    // main handler as iafon.HandlerFunc
    s.GET("/handler/:param_name", func (c *iafon.Context) {
        // Now, iafon.Context has the following field
        // type Context struct {
        //     Rsp http.ResponseWriter
        //     Req *http.Request
        //
        //     // parameter in route pattern
        //     Param map[string]string
        //
        //     // we can put any thing in this map for passing through middlewares and main handler
        //     Udata map[string]interface{}  
        // }
        fmt.Fprintf(c.Rsp, "Hello from iafon.HandlerFunc. param: %s\n", c.Param["param_name"])
    })

    // main handler as http.Handler
    // both "/handler" and "/handler/" could be handled
    // if only handle "/handler/", "/handler" will auto redirect to "/handler/"
    s.GET("/handler/", &HttpHandler{})

    // main handler as http.HandlerFunc
    // "/handler/http" has higher priority than "/handler/:param_name" 
    s.GET("/handler/http", func (w http.ResponseWriter, r *http.Request) {
        fmt.Fprint(w, "Hello from http.HandlerFunc\n")
    })

    // main handler as controller method
    // "/user" will redirect to "/user/"
    s.GET("/user/", (*AController).Index)

    // equivalent to s.Handle("GET", "/user/:id", (*AController).Show)
    s.GET("/user/:id", (*AController).Show)

    s.POST("/user/", (*AController).Store)
    s.PUT("/user/:id", (*AController).Update)
    s.DELETE("/user/:id", (*AController).Destroy)

    // handle request to this route in specified http methods
    s.Some([]string{"POST", "PUT"}, "/user/test", func (c *iafon.Context) {
        fmt.Fprint(c.Rsp, "Hello from handle some\n")
    })

    // handle request to this route in any http method
    // equivalent to s.Handle("*", "/", func (c *iafon.Context) {})
    s.Any("/", func (c *iafon.Context) {
        fmt.Fprint(c.Rsp, "Hello from handle any\n")
    })

    // create a sub group of routes
    // using group, we can set group prefix and middlewares
    {
        // s on left hand of ':=' is a new variable in this sub scope
        // the first parameter of Group method is group prefix
        s := s.Group("/admin")

        // this middleware is use to routes in this group
        s.UseMiddleware(&CMiddleware{})

        // the full path will be "/admin/user/"
        s.GET("/user/", (*BController).Index)

        // the full path will be "/admin/user/:id"
        s.GET("/user/:id", (*BController).Show)
        
        s.POST("/user/", (*BController).Store)
        s.PUT("/user/:id", (*BController).Update)
    }

    {
        // WAINING: if route pattern is not start with '/',
        // the first path node "x.org" will be treat as host name
        s := s.Group("x.org/admin")

        // if we request "http://x.org/admin/user/", (*CController).Index will handle the request.
        // if we request "http://y.org/admin/user/", (*BController).Index will handle the request.
        s.GET("/user/", (*CController).Index)

        rn := s.DELETE("/user/:id", (*CController).Destroy)

        // a single route could also use middleware
        // if we want middleware to be executed after route main handler,
        // then we should provide the second parameter as negative integer
        rn.UseMiddleware(&DMiddleware{}, -1)
    }

    // we could customize the following three http error handler

    // route not found
    // error handler as iafon.Handler
    s.HandleError(404, &ErrorHandler404{})

    // route found but http method is not allowed
    // error handler as iafon.HandlerFunc
    s.HandleError(405, func (c *iafon.Context) {
        http.Error(c.Rsp, "405 method not allowed", 405)
    })

    // 500 means server panic when handle request
    s.HandleError(500, func (c *iafon.Context) {
        http.Error(c.Rsp, "500 internal server error", 500)
    })

    // let's print all routes added
    fmt.Println(s.GetRoutes().String())

    // after we finish all routing config, run the server
    s.Run()

    // we could create multiple server, then use iafon.RunServers or iafon.RunServersWaitAll to run servers
    // s1 := iafon.NewServer(":8091")
    // s1.GET("/", func(*iafon.Context){})
    // s2 := iafon.NewServer(":8092")
    // s2.GET("/", func(*iafon.Context){})
    // iafon.RunServersWaitAll(s, s1, s2)
}

type AMiddleware struct {
    iafon.Middleware
}

func (m *AMiddleware) Handle() bool {
    // if a Middleware return false, the middlewares and main handler after this middleware will not be executed
    return true
}

type BMiddleware struct {
    iafon.Middleware
}

func (m *BMiddleware) Handle() bool {
    return true
}

type CMiddleware struct {
    iafon.Middleware
}

func (m *CMiddleware) Handle() bool {
    return true
}

type DMiddleware struct {
    iafon.Middleware
}

func (m *DMiddleware) Handle() bool {
    return true
}

type IafonHandler struct {}

func (h *IafonHandler) Handle(c *iafon.Context) {
    fmt.Fprint(c.Rsp, "Hello from iafon.Handler\n")
}

type ErrorHandler404 struct {}

func (h *ErrorHandler404) Handle(c *iafon.Context) {
    http.Error(c.Rsp, "404 route not found", 404)
}

type HttpHandler struct {}

func (h *HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "Hello from http.Handler\n")
}

func init() {
    // controller must be registerd before using its method as route handler
    iafon.RegisterController(&AController{})
    iafon.RegisterController(&BController{})
    iafon.RegisterController(&CController{})
}

type AController struct {
    // comstom controller should embed iafon.Controller
    // iafon.Context is embeded in iafon.Controller
    // so we can use Controller instance as Context instance
    iafon.Controller
}

func (c *AController) Initialize() {
    // Initialize method will be executed before any controller method added as main route handler
    fmt.Fprint(c.Rsp, "Hello from (*AController).Initialize\n")
}

func (c *AController) Finalize() {
    // Finalize method will be executed after any controller method added as main route handler
    fmt.Fprint(c.Rsp, "Hello from (*AController).Finalize\n")
}

// before executing Index, Initialize is executed
// after executing Index, Finalize is executed
func (c *AController) Index() {
    fmt.Fprint(c.Rsp, "Hello from (*AController).Index\n")
}

// before executing Show, Initialize is executed
// after executing Show, Finalize is executed
func (c *AController) Show() {
    fmt.Fprintf(c.Rsp, "Hello from (*AController).Show. id: %s\n", c.Param["id"])
}

func (c *AController) Store() {
    fmt.Fprintf(c.Rsp, "Hello from (*AController).Store\n")
}

func (c *AController) Update() {
    fmt.Fprintf(c.Rsp, "Hello from (*AController).Update. id: %s\n", c.Param["id"])
}

func (c *AController) Destroy() {
    fmt.Fprintf(c.Rsp, "Hello from (*AController).Destroy. id: %s\n", c.Param["id"])
}

type BController struct {
    iafon.Controller
}

func (c *BController) Index() {
    fmt.Fprint(c.Rsp, "Hello from (*BController).Index\n")
}

func (c *BController) Show() {
    fmt.Fprintf(c.Rsp, "Hello from (*BController).Show. id: %s\n", c.Param["id"])
}

func (c *BController) Store() {
    fmt.Fprintf(c.Rsp, "Hello from (*BController).Store\n")
}

func (c *BController) Update() {
    fmt.Fprintf(c.Rsp, "Hello from (*BController).Update. id: %s\n", c.Param["id"])
}

func (c *BController) Destroy() {
    fmt.Fprintf(c.Rsp, "Hello from (*BController).Destroy. id: %s\n", c.Param["id"])
}

type CController struct {
    iafon.Controller
}

func (c *CController) Index() {
    fmt.Fprint(c.Rsp, "Hello from (*CController).Index\n")
}

func (c *CController) Show() {
    fmt.Fprintf(c.Rsp, "Hello from (*CController).Show. id: %s\n", c.Param["id"])
}

func (c *CController) Store() {
    fmt.Fprintf(c.Rsp, "Hello from (*CController).Store\n")
}

func (c *CController) Update() {
    fmt.Fprintf(c.Rsp, "Hello from (*CController).Update. id: %s\n", c.Param["id"])
}

func (c *CController) Destroy() {
    fmt.Fprintf(c.Rsp, "Hello from (*CController).Destroy. id: %s\n", c.Param["id"])
}
```
