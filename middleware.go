package iafon

import (
	"reflect"
)

type MiddlewareInterface interface {
	GetBaseMiddleware() *Middleware
	RealExecOrder() int
	Handle() bool
}

var middlewareAddOrder int16

// map from MiddlewareInterface.RealExecOrder to *reflect.Value
var addedMiddlewares = make(map[int]*reflect.Value)

func markMiddlewareAsAdded(m MiddlewareInterface, execOrder ...int16) {
	middlewareAddOrder++

	baseMiddleware := m.GetBaseMiddleware()
	baseMiddleware.addOrder = middlewareAddOrder

	if len(execOrder) > 0 {
		baseMiddleware.ExecOrder = execOrder[0]
	}

	value := reflect.Indirect(reflect.ValueOf(m))
	addedMiddlewares[m.RealExecOrder()] = &value
}

// custom middleware should embed base Middleware
type Middleware struct {
	*Context

	ExecOrder int16
	addOrder  int16
}

func (m *Middleware) GetBaseMiddleware() *Middleware {
	return m
}

func (m *Middleware) RealExecOrder() int {
	execOrder := m.ExecOrder
	if execOrder >= 0 {
		execOrder += 1
	}
	return (int(execOrder) << 16) - int(m.addOrder)
}

func (m *Middleware) Handle() bool {
	return true
}
