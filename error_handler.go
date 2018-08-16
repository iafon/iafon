package iafon

import (
	"fmt"
	"net/http"
)

func handleErrorByDefault(code int, c *Context) {
	switch code {
	case 404:
		http.Error(c.Rsp, "404 route not found", code)
	case 405:
		http.Error(c.Rsp, "405 method not allowed", code)
	case 500:
		http.Error(c.Rsp, "500 internal server error", code)
	default:
		fmt.Printf("http error %d is not handled", code)
	}
}
