package ws

import (
	"log"
	"strings"
)

func init() {
	var tree = new(node)

	tree.addRoute("/a/:name", &Hba{})

	log.Println(tree.getValue("/a/1"))
}

type GroupFunction func()

type HttpFunction func(t *Stream)

type Before func(t *Stream) (interface{}, error)

type After func(t *Stream) error

type Http struct {
	IgnoreCase bool
	// Routers    map[string]map[string]*Hba
	Router *node
}

type Hba struct {
	Method  string
	Handler HttpFunction
	Before  []Before
	After   []After
}

var globalHttpPath string
var globalBefore []Before
var globalAfter []After

const (
	PassBefore uint8 = 1 << iota
	PassAfter
	ForceBefore
	ForceAfter
)

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

	var m = strings.ToUpper(method)

	path = globalHttpPath + path

	if h.IgnoreCase {
		path = strings.ToUpper(path)
	}

	if h.Router == nil {
		h.Router = new(node)
	}
	//
	// if _, ok := h.Routers[m]; !ok {
	// 	h.Routers[m] = make(map[string]*Hba)
	// }
	//
	// if h.Routers[m][path] != nil {
	// 	println(m, path, "already set route")
	// 	return
	// }

	var hba = &Hba{}

	var handler HttpFunction
	var before []Before
	var after []After

	var passBefore = false
	var passAfter = false
	var forceBefore = false
	var forceAfter = false

	for _, mark := range v {
		switch mark.(type) {
		case uint8:
			if mark.(uint8) == PassBefore {
				passBefore = true
			}
			if mark.(uint8) == PassAfter {
				passAfter = true
			}
			if mark.(uint8) == ForceBefore {
				forceBefore = true
			}
			if mark.(uint8) == ForceAfter {
				forceAfter = true
			}
		}
	}

	for _, fn := range v {
		switch fn.(type) {
		case func(t *Stream):
			handler = fn.(func(t *Stream))
		case []Before:
			before = fn.([]Before)
		case []After:
			after = fn.([]After)
		}
	}

	if handler == nil {
		println(m, path, "handler function is nil")
		return
	}

	hba.Handler = handler

	hba.Before = append(globalBefore, before...)
	if passBefore {
		hba.Before = nil
	}
	if forceBefore {
		hba.Before = before
	}

	hba.After = append(globalAfter, after...)
	if passAfter {
		hba.After = nil
	}
	if forceAfter {
		hba.After = after
	}

	// h.Routers[m][path] = hba

	hba.Method = m

	h.Router.addRoute(path, hba)
}

func (h *Http) GetRoute(method string, path string) (*Hba, Params) {

	if h.IgnoreCase {
		path = strings.ToUpper(path)
	}

	if h.Router == nil {
		return nil, nil
	}

	// var m = strings.ToUpper(method)

	// if f, ok := h.Routers[m][path]; ok {
	// 	return f
	// }

	handle, p, _ := h.Router.getValue(path)

	return handle, p
}

func (h *Http) Get(path string, v ...interface{}) {
	h.SetRoute("GET", path, v...)
}

func (h *Http) Post(path string, v ...interface{}) {
	h.SetRoute("POST", path, v...)
}
