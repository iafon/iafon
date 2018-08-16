package iafon

import (
	"fmt"
	"sort"
	"strconv"
)

type RouteNode struct {
	host     string
	method   string
	pattern  string
	handlers []*tMixHandler
}

func (rn *RouteNode) UseMiddleware(m MiddlewareInterface, execOrder ...int16) *RouteNode {
	if m.GetBaseMiddleware().addOrder == 0 {
		markMiddlewareAsAdded(m, execOrder...)
	}

	h := newMixHandler(m)

	i := len(rn.handlers) - 1

	for ; i >= 0; i-- {
		if rn.handlers[i].order > h.order {
			break
		}
	}

	rn.handlers = append(rn.handlers, nil)
	copy(rn.handlers[i+2:], rn.handlers[i+1:])
	rn.handlers[i+1] = h

	return rn
}

func newRouteNode(host, method, pattern string, mainHandler interface{}) *RouteNode {
	rn := &RouteNode{
		host:     host,
		method:   method,
		pattern:  pattern,
		handlers: []*tMixHandler{newMixHandler(mainHandler)},
	}
	if rn.handlers[0].hType == cHTYPE_MIDDLEWARE {
		panic("middleware can not be used as route main handler.")
	}
	return rn
}

type RouteNodeSlice []*RouteNode

func (s RouteNodeSlice) Len() int {
	return len(s)
}

func (s RouteNodeSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s RouteNodeSlice) Less(i, j int) bool {
	a, b := s[i], s[j]
	if a.host < b.host {
		return true
	} else if a.host == b.host {
		if a.pattern < b.pattern {
			return true
		} else if a.pattern == b.pattern {
			if a.method < b.method {
				return true
			} else {
				return false
			}
		} else {
			return false
		}
	} else {
		return false
	}
}

func (s RouteNodeSlice) Sort() RouteNodeSlice {
	sort.Sort(s)
	return s
}

func (s RouteNodeSlice) String() string {
	str := ""

	max_host_len := 0
	max_method_len := 0
	for _, rn := range s {
		if len(rn.host) > max_host_len {
			max_host_len = len(rn.host)
		}
		if len(rn.method) > max_method_len {
			max_method_len = len(rn.method)
		}
	}

	for _, rn := range s {
		str += fmt.Sprintf("%-"+strconv.Itoa(max_method_len)+"s %-"+strconv.Itoa(max_host_len)+"s %s\n", rn.method, rn.host, rn.pattern)
	}

	return str
}

func (s RouteNodeSlice) Print() {
	fmt.Println(s.String())
}
