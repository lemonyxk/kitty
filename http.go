package ws

type HttpFunction func(t *Stream)

type HttpMiddle func(t *Stream) (interface{}, error)

type HttpHandle struct {
	Middle  HttpMiddle
	Routers map[string]map[string]HttpFunction
}

func (h *HttpHandle) Init() {
	h.Routers = make(map[string]map[string]HttpFunction)
}

func (h *HttpHandle) HSetRoute(method string, router string, f HttpFunction) {
	if _, ok := h.Routers[method]; !ok {
		h.Routers[method] = make(map[string]HttpFunction)
	}
	h.Routers[method][router] = f
}

func (h *HttpHandle) HGetRoute(method string, router string) HttpFunction {
	if f, ok := h.Routers[method][router]; ok {
		return f
	}
	return nil
}
