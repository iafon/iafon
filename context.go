package iafon

import (
	"net/http"
)

type Context struct {
	Rsp   http.ResponseWriter
	Req   *http.Request
	Param map[string]string
	Udata map[string]interface{}
}
