package iafon

import (
	"net/http"
	"testing"
)

var controller_test_echo string

type TestController struct {
	Controller
}

func (t *TestController) Index() {
	controller_test_echo = t.Req.URL.Path
}

func TestRegisterController(t *testing.T) {
	RegisterController(&TestController{})

	value, ok := registeredControllers["iafon.TestController"]

	if !ok {
		t.Fatal("RegisterController failed")
	}

	if value.Type().String() != "iafon.TestController" {
		t.Fatal("RegisterController as invalid type")
	}
}

func TestHandleController(t *testing.T) {
	s := NewServer("127.0.0.1:")

	method := "GET"
	path := "/Controller"

	rn := s.Handle(method, path, (*TestController).Index)
	if rn == nil {
		t.Fatal("fail to add Controller method as handler")
	}

	req, _ := http.NewRequest(method, "http://localhost"+path, nil)

	s.ServeHTTP(nil, req)

	if controller_test_echo != path {
		t.Fatal("Controller method is not fired on request")
	}
}
