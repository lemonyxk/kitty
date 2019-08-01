package ws

import "strings"

type HttpFunction func(t *Stream)

type HttpMiddle func(t *Stream) (interface{}, error)

var globalHttpPath = ""

type HttpHandle struct {
	Middle  HttpMiddle
	Routers map[string]map[string]HttpFunction
}

func (h *HttpHandle) Group(path string, fn func()) {
	globalHttpPath = path
	fn()
	globalHttpPath = ""
}

func (h *HttpHandle) SetRoute(method string, path string, f HttpFunction) {

	path = globalHttpPath + path

	if h.Routers == nil {
		h.Routers = make(map[string]map[string]HttpFunction)
	}

	var m = strings.ToUpper(method)

	if _, ok := h.Routers[m]; !ok {
		h.Routers[m] = make(map[string]HttpFunction)
	}

	h.Routers[m][path] = f
}

func (h *HttpHandle) GetRoute(method string, path string) HttpFunction {

	if h.Routers == nil {
		h.Routers = make(map[string]map[string]HttpFunction)
	}

	var m = strings.ToUpper(method)

	if f, ok := h.Routers[m][path]; ok {
		return f
	}

	return nil
}

func (h *HttpHandle) Get(path string, f HttpFunction) {
	h.SetRoute("GET", path, f)
}

func (h *HttpHandle) Post(path string, f HttpFunction) {
	h.SetRoute("POST", path, f)
}
