package iafon

import (
	"fmt"
	"net"
	"net/http"
	"path"
	"runtime"
	"strings"
)

var http_methods = map[string]bool{
	"*":       true, // Any
	"GET":     true,
	"POST":    true,
	"PUT":     true,
	"DELETE":  true,
	"OPTIONS": true,
	"HEAD":    true,
	"PATCH":   true,
	"CONNECT": true,
	"TRACE":   true,
}

type tMap_Method_RouteNode map[string]*RouteNode

type tMap_Host_Method_RouteNode struct {
	hosts           map[string]tMap_Method_RouteNode
	shouldMatchHost bool
}

type Router struct {
	RouteGroup
	matcher       *PatternMapByTree
	errorHandlers map[int]Handler
}

func newRouter() *Router {
	r := &Router{}
	r.RouteGroup.router = r
	r.matcher = &PatternMapByTree{}
	return r
}

func (r *Router) HandleError(code int, handler interface{}) {
	if code != 404 && code != 405 && code != 500 {
		panic("HandleError only support 404 405 500 http code")
	}

	if r.errorHandlers == nil {
		r.errorHandlers = make(map[int]Handler)
	}

	switch h := handler.(type) {
	case Handler:
		r.errorHandlers[code] = h
	case func(*Context):
		r.errorHandlers[code] = HandlerFunc(h)
	default:
		panic("invalid http error handler type.")
	}
}

func (r *Router) addRoute(method, pattern string, handler interface{}) *RouteNode {
	method = strings.ToUpper(method)

	if !http_methods[method] {
		panic("http: invalid method " + method)
	}
	if len(pattern) == 0 {
		panic("http: route pattern can not be empty")
	}
	if handler == nil {
		panic("http: nil handler")
	}

	raw_pattern := pattern
	host := ""

	if pattern[0] != '/' {
		if pos := strings.IndexByte(pattern, '/'); pos > 0 {
			host = pattern[:pos]
			pattern = pattern[pos:]
		} else {
			panic(fmt.Sprintf("http: invalid pattern, you need /%s or %[1]s/, meaning /path, host/path", pattern))
		}
	}

	var m *tMap_Host_Method_RouteNode

	if v := r.matcher.Get(pattern); v != nil {
		m = v.(*tMap_Host_Method_RouteNode)
	}

	if m == nil {
		m = &tMap_Host_Method_RouteNode{}
		m.hosts = make(map[string]tMap_Method_RouteNode)
	} else if m.hosts[host] != nil && m.hosts[host][method] != nil {
		panic(fmt.Sprintf("http: duplicate route '%s %s'", method, raw_pattern))
	}

	if m.hosts[host] == nil {
		m.hosts[host] = make(tMap_Method_RouteNode)
	}

	if host != "" {
		m.shouldMatchHost = true
	}

	rn := newRouteNode(host, method, pattern, handler)

	m.hosts[host][method] = rn

	r.matcher.Set(pattern, m)

	return rn
}

// implement http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := &Context{Rsp: w, Req: req}

	defer func() {
		if p := recover(); p != nil {
			fmt.Println("internal server error:", p)
			fmt.Println("stack trace:", stackTrace(false))

			r.handleError(500, ctx)
		}
	}()

	if req.RequestURI == "*" {
		if req.ProtoAtLeast(1, 1) {
			w.Header().Set("Connection", "close")
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	host := req.URL.Host
	path := req.URL.Path

	// CONNECT requests are not canonicalized.
	if req.Method != "CONNECT" {
		host = stripHostPort(req.Host)
		path = cleanPath(req.URL.Path)
	}

	v, params, redirect, _ := r.matcher.Match(path)
	if v == nil {
		r.handleError(404, ctx)
		return
	}

	ctx.Param = params

	m := v.(*tMap_Host_Method_RouteNode)

	var rn *RouteNode

	hostMatched := false

	// Host-specific pattern takes precedence over generic ones
	if m.shouldMatchHost {
		if mh := m.hosts[host]; mh != nil {
			hostMatched = true

			rn = mh[req.Method]
			if rn == nil {
				// any method
				rn = mh["*"]
			}
		}
	}

	// if shouldMatchHost, rn may set already
	if rn == nil {
		if mh := m.hosts[""]; mh != nil {
			hostMatched = true

			rn = mh[req.Method]
			if rn == nil {
				// any method
				rn = mh["*"]
			}
		}
	}

	if !hostMatched {
		r.handleError(404, ctx)
		return
	}

	if rn == nil {
		r.handleError(405, ctx)
		return
	}

	// redirect if path not match exactly
	// 1. redirect "/path" to "/path/" if "/path" is not registered but "/path/" is registered
	// 2. redirect "/path//sub" "/path///sub" to "/path/sub"
	// 3. redirect "/path/sub/.." to "/path"
	// 4. redirect "/path/sub/." to "/path/sub"
	// ...
	if redirect || path != req.URL.Path {
		url := *req.URL
		if redirect {
			url.Path = path + "/"
		} else {
			url.Path = path
		}
		http.Redirect(w, req, url.String(), http.StatusTemporaryRedirect)
		return
	}

	for _, h := range rn.handlers {
		if next := h.call(ctx); !next {
			break
		}
	}
}

func (r *Router) handleError(code int, ctx *Context) {
	if h := r.errorHandlers[code]; h != nil {
		h.Handle(ctx)
	} else {
		handleErrorByDefault(code, ctx)
	}
}

// this function is from github.com/golang/go/src/net/http/server.go
// stripHostPort returns h without any trailing ":<port>".
func stripHostPort(h string) string {
	// If no port on host, return unchanged
	if strings.IndexByte(h, ':') == -1 {
		return h
	}
	host, _, err := net.SplitHostPort(h)
	if err != nil {
		return h // on error, return unchanged
	}
	return host
}

// this function is from github.com/golang/go/src/net/http/server.go
// cleanPath returns the canonical path for p, eliminating . and .. elements.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		// Fast path for common case of p being the string we want:
		if len(p) == len(np)+1 && strings.HasPrefix(p, np) {
			np = p
		} else {
			np += "/"
		}
	}
	return np
}

// this function is from http://theo.im/blog/2014/07/21/Printing-stacktrace-in-Go/
func stackTrace(_all ...bool) string {
	all := false
	if len(_all) > 0 {
		all = _all[0]
	}

	// Reserve 10K buffer at first
	buf := make([]byte, 1024*2)
	for {
		size := runtime.Stack(buf, all)
		// The size of the buffer may be not enough to hold the stacktrace,
		// so double the buffer size
		if size == len(buf) {
			buf = make([]byte, len(buf)<<1)
			continue
		}
		break
	}
	return string(buf)
}
