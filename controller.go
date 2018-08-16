package iafon

import (
	"reflect"
)

type ControllerInterface interface {
	GetBaseController() *Controller
	Initialize()
	Finalize()
}

type Controller struct {
	*Context
}

func (c *Controller) GetBaseController() *Controller {
	return c
}

func (c *Controller) Initialize() {}

func (c *Controller) Finalize() {}

var registeredControllers = make(map[string]*reflect.Value)

func RegisterController(c ControllerInterface) {
	value := reflect.Indirect(reflect.ValueOf(c))
	registeredControllers[value.Type().String()] = &value
}
