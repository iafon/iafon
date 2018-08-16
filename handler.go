package iafon

type Handler interface {
	Handle(*Context)
}

type HandlerFunc func(*Context)

func (h HandlerFunc) Handle(c *Context) {
	h(c)
}
