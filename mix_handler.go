package iafon

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

type tMixHandlerType int8

const (
	cHTYPE_HTTP_HANDLER tMixHandlerType = iota
	cHTYPE_IAFON_HANDLER
	cHTYPE_CONTROLLER
	cHTYPE_MIDDLEWARE
)

type tMixHandler struct {
	hType tMixHandlerType

	httpHandler  http.Handler
	iafonHandler Handler

	controllerMethod reflect.Value
	controllerValue  *reflect.Value

	middlewareValue *reflect.Value

	// execution order
	order int
}

func newMixHandler(handler interface{}) *tMixHandler {
	h := &tMixHandler{}

	h.order = 0

	switch handler := handler.(type) {
	case http.Handler:
		h.hType = cHTYPE_HTTP_HANDLER
		h.httpHandler = handler
	case http.HandlerFunc:
		h.hType = cHTYPE_HTTP_HANDLER
		h.httpHandler = handler
	case func(http.ResponseWriter, *http.Request):
		h.hType = cHTYPE_HTTP_HANDLER
		h.httpHandler = http.HandlerFunc(handler)

	case Handler:
		h.hType = cHTYPE_IAFON_HANDLER
		h.iafonHandler = handler
	case HandlerFunc:
		h.hType = cHTYPE_IAFON_HANDLER
		h.iafonHandler = handler
	case func(*Context):
		h.hType = cHTYPE_IAFON_HANDLER
		h.iafonHandler = HandlerFunc(handler)

	case MiddlewareInterface:
		h.order = handler.RealExecOrder()
		h.hType = cHTYPE_MIDDLEWARE
		h.middlewareValue = addedMiddlewares[h.order]
	default:
		// controller method
		t := fmt.Sprintf("%T", handler)

		panic_msg := "invalid handler type: " + t

		if t[:5] != "func(" {
			panic(panic_msg)
		}

		if t[5] == '*' {
			t = t[6:]
		} else {
			t = t[5:]
		}

		pos := strings.Index(t, ",")
		if pos >= 0 {
			panic(panic_msg)
		}

		pos = strings.Index(t, ")")

		controllerTypeName := t[:pos]

		if controllerTypeName == "" {
			panic(panic_msg)
		}

		if _, ok := registeredControllers[controllerTypeName]; !ok {
			panic("not registered controller type: " + controllerTypeName)
		}

		h.hType = cHTYPE_CONTROLLER
		h.controllerMethod = reflect.ValueOf(handler)
		h.controllerValue = registeredControllers[controllerTypeName]
	}

	return h
}

func (h *tMixHandler) call(ctx *Context) bool {
	var next = true
	switch h.hType {
	case cHTYPE_HTTP_HANDLER:
		h.httpHandler.ServeHTTP(ctx.Rsp, ctx.Req)
	case cHTYPE_IAFON_HANDLER:
		h.iafonHandler.Handle(ctx)
	case cHTYPE_MIDDLEWARE:
		mValue := reflect.New(h.middlewareValue.Type())
		mValue.Elem().Set(*h.middlewareValue)

		m := mValue.Interface().(MiddlewareInterface)
		m.GetBaseMiddleware().Context = ctx

		next = m.Handle()
	case cHTYPE_CONTROLLER:
		cValue := reflect.New(h.controllerValue.Type())
		cValue.Elem().Set(*h.controllerValue)

		c := cValue.Interface().(ControllerInterface)
		c.GetBaseController().Context = ctx

		c.Initialize()
		h.controllerMethod.Call([]reflect.Value{cValue})
		c.Finalize()
	default:
		panic("invalid handler type")
	}
	return next
}
