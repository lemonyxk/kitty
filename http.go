package ws

import "strings"

type HttpFunction func(t *Stream)

type HttpMiddle func(t *Stream) (interface{}, error)

type HttpHandle struct {
	Middle  HttpMiddle
	Routers map[string]map[string]HttpFunction
}

func (h *HttpHandle) HSetRoute(method string, router string, f HttpFunction) {

	if h.Routers == nil {
		h.Routers = make(map[string]map[string]HttpFunction)
	}

	var m = strings.ToUpper(method)

	if _, ok := h.Routers[m]; !ok {
		h.Routers[m] = make(map[string]HttpFunction)
	}

	h.Routers[m][router] = f
}

func (h *HttpHandle) HGetRoute(method string, router string) HttpFunction {

	if h.Routers == nil {
		h.Routers = make(map[string]map[string]HttpFunction)
	}

	var m = strings.ToUpper(method)

	if f, ok := h.Routers[m][router]; ok {
		return f
	}

	return nil
}
