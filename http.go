package ws

import "strings"

type GroupFunction func()

type HttpFunction func(t *Stream)

type Before func(t *Stream) (interface{}, error)

type After func(t *Stream) error

type HttpHandle struct {
	Routers map[string]map[string]*hba
}

type hba struct {
	Handler HttpFunction
	Before  []Before
	After   []After
}

var globalHttpPath string
var globalBefore []Before
var globalAfter []After

func (h *HttpHandle) Group(path string, v ...interface{}) {

	if v == nil {
		panic("Group function length is 0")
	}

	var g GroupFunction

	for _, fn := range v {
		switch fn.(type) {
		case GroupFunction:
			g = fn.(GroupFunction)
		case []Before:
			globalBefore = fn.([]Before)
		case []After:
			globalAfter = fn.([]After)
		}
	}

	if g == nil {
		panic("Group function is nil")
	}

	globalHttpPath = path
	g()
	globalHttpPath = ""
	globalBefore = nil
	globalAfter = nil
}

func (h *HttpHandle) SetRoute(method string, path string, v ...interface{}) {

	path = globalHttpPath + path

	if h.Routers == nil {
		h.Routers = make(map[string]map[string]*hba)
	}

	var m = strings.ToUpper(method)

	if _, ok := h.Routers[m]; !ok {
		h.Routers[m] = make(map[string]*hba)
	}

	if h.Routers[m][path] != nil {
		println(m, path, "already set route")
		return
	}

	var hba = &hba{}

	for _, fn := range v {
		switch fn.(type) {
		case HttpFunction:
			hba.Handler = fn.(HttpFunction)
		case []Before:
			if globalBefore != nil {
				hba.Before = globalBefore
			}
			hba.Before = append(hba.Before, fn.([]Before)...)
		case []After:
			if globalAfter != nil {
				hba.After = globalAfter
			}
			hba.After = append(hba.After, fn.([]After)...)
		}
	}

	h.Routers[m][path] = hba
}

func (h *HttpHandle) GetRoute(method string, path string) *hba {

	if h.Routers == nil {
		h.Routers = make(map[string]map[string]*hba)
	}

	var m = strings.ToUpper(method)

	if f, ok := h.Routers[m][path]; ok {
		return f
	}

	return nil
}

func (h *HttpHandle) Get(path string, v ...interface{}) {
	h.SetRoute("GET", path, v)
}

func (h *HttpHandle) Post(path string, v ...interface{}) {
	h.SetRoute("POST", path, v)
}
