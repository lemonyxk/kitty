package ws

import "strings"

type GroupFunction func()

type HttpFunction func(t *Stream)

type Before func(t *Stream) (interface{}, error)

type After func(t *Stream) error

type Http struct {
	IgnoreCase bool
	Routers    map[string]map[string]*hba
}

type hba struct {
	Handler HttpFunction
	Before  []Before
	After   []After
}

var globalHttpPath string
var globalBefore []Before
var globalAfter []After

func (h *Http) Group(path string, v ...interface{}) {

	if v == nil {
		panic("Group function length is 0")
	}

	var g GroupFunction

	for _, fn := range v {
		switch fn.(type) {
		case func():
			g = fn.(func())
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

func (h *Http) SetRoute(method string, path string, v ...interface{}) {

	path = globalHttpPath + path

	if h.IgnoreCase {
		path = strings.ToUpper(path)
	}

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
	hba.Before = globalBefore
	hba.After = globalAfter

	for _, fn := range v {
		switch fn.(type) {
		case func(t *Stream):
			hba.Handler = fn.(func(t *Stream))
		case []Before:
			hba.Before = append(hba.Before, fn.([]Before)...)
		case []After:
			hba.After = append(hba.After, fn.([]After)...)
		}
	}

	h.Routers[m][path] = hba
}

func (h *Http) GetRoute(method string, path string) *hba {

	if h.Routers == nil {
		return nil
	}

	var m = strings.ToUpper(method)

	if f, ok := h.Routers[m][path]; ok {
		return f
	}

	return nil
}

func (h *Http) Get(path string, v ...interface{}) {
	h.SetRoute("GET", path, v...)
}

func (h *Http) Post(path string, v ...interface{}) {
	h.SetRoute("POST", path, v...)
}
