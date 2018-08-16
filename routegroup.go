package iafon

import (
	"fmt"
)

type RouteGroup struct {
	router *Router
	parent *RouteGroup

	routes    []*RouteNode
	subgroups []*RouteGroup

	prefix      string
	middlewares []MiddlewareInterface
}

func (g *RouteGroup) SetPrefix(prefix string) *RouteGroup {
	if len(g.routes) > 0 {
		panic("prefix should be set before adding routes")
	}
	if len(g.subgroups) > 0 {
		panic("prefix should be set before adding sub groups")
	}
	if g.parent != nil && g.parent.prefix != "" {
		if g.parent.prefix[len(g.parent.prefix)-1] == '/' && prefix[0] == '/' {
			panic(fmt.Sprintf("prefix '%s' concat '%s' to form invalid path.", g.parent.prefix, prefix))
		}
		g.prefix = g.parent.prefix + prefix
	} else {
		g.prefix = prefix
	}
	return g
}

func (g *RouteGroup) UseMiddleware(m MiddlewareInterface, execOrder ...int16) *RouteGroup {
	if m.GetBaseMiddleware().addOrder == 0 {
		markMiddlewareAsAdded(m, execOrder...)
	}

	for _, r := range g.routes {
		r.UseMiddleware(m)
	}

	for _, subgroup := range g.subgroups {
		subgroup.UseMiddleware(m)
	}

	g.middlewares = append(g.middlewares, m)

	return g
}

func (g *RouteGroup) Handle(method, pattern string, handler interface{}) *RouteNode {
	pattern = g.prefix + pattern

	rn := g.router.addRoute(method, pattern, handler)

	for _, m := range g.middlewares {
		rn.UseMiddleware(m)
	}

	g.routes = append(g.routes, rn)

	return rn
}

func (g *RouteGroup) GET(pattern string, handler interface{}) *RouteNode {
	return g.Handle("GET", pattern, handler)
}

func (g *RouteGroup) POST(pattern string, handler interface{}) *RouteNode {
	return g.Handle("POST", pattern, handler)
}

func (g *RouteGroup) PUT(pattern string, handler interface{}) *RouteNode {
	return g.Handle("PUT", pattern, handler)
}

func (g *RouteGroup) DELETE(pattern string, handler interface{}) *RouteNode {
	return g.Handle("DELETE", pattern, handler)
}

func (g *RouteGroup) OPTIONS(pattern string, handler interface{}) *RouteNode {
	return g.Handle("OPTIONS", pattern, handler)
}

func (g *RouteGroup) HEAD(pattern string, handler interface{}) *RouteNode {
	return g.Handle("HEAD", pattern, handler)
}

func (g *RouteGroup) PATCH(pattern string, handler interface{}) *RouteNode {
	return g.Handle("PATCH", pattern, handler)
}

func (g *RouteGroup) Any(pattern string, handler interface{}) *RouteNode {
	return g.Handle("*", pattern, handler)
}

func (g *RouteGroup) Some(methods []string, pattern string, handler interface{}) *RouteGroup {
	return g.Group(func(group *RouteGroup) {
		for _, method := range methods {
			group.Handle(method, pattern, handler)
		}
	})
}

func (g *RouteGroup) Group(p ...interface{}) *RouteGroup {
	subgroup := &RouteGroup{router: g.router, parent: g, prefix: g.prefix}
	for _, m := range g.middlewares {
		subgroup.UseMiddleware(m)
	}
	g.subgroups = append(g.subgroups, subgroup)

	for i, v := range p {
		switch v := v.(type) {
		case string:
			if i != 0 {
				panic("route group prefix should be the first parameter.")
			}
			subgroup.SetPrefix(v)
		case MiddlewareInterface:
			subgroup.UseMiddleware(v)
		case func(*RouteGroup):
			if i != len(p)-1 {
				panic("route group callback should be the last parameter.")
			}
			v(subgroup)
		default:
			panic(fmt.Sprintf("%dth parameter is invalid for Group method. parameter type %T", i, v))
		}
	}

	return subgroup
}

func (g *RouteGroup) GetRoutes() RouteNodeSlice {
	routes := append(RouteNodeSlice{}, g.routes...)
	for _, subgroup := range g.subgroups {
		routes = append(routes, subgroup.GetRoutes()...)
	}
	routes.Sort()
	return routes
}
