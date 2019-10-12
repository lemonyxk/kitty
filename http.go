package lemo

import (
	"net/http"
	"strings"

	"github.com/Lemo-yxk/tire"
)

type HttpGroupFunction func()

type StreamFunction func(t *Stream) func() *Error

type HttpFunction func(w http.ResponseWriter, r *http.Request)

type HttpBefore func(t *Stream) (Context, func() *Error)

type HttpAfter func(t *Stream) func() *Error

type ErrorFunction func(func() *Error)

type Http struct {
	IgnoreCase bool
	Router     *tire.Tire
	OnError    ErrorFunction
}

type HBA struct {
	Path           []byte
	Route          []byte
	Method         string
	StreamFunction StreamFunction
	HttpFunction   HttpFunction
	Before         []HttpBefore
	After          []HttpAfter
}

var globalHttpPath string
var globalHttpBefore []HttpBefore
var globalHttpAfter []HttpAfter

const (
	PassBefore uint8 = 1 << iota
	PassAfter
	ForceBefore
	ForceAfter
)

func (h *Http) Ready() {

}

func (h *Http) Group(path string, v ...interface{}) {

	if v == nil {
		panic("Group function length is 0")
	}

	var g HttpGroupFunction

	for _, fn := range v {
		switch fn.(type) {
		case func():
			g = fn.(func())
		case []HttpBefore:
			globalHttpBefore = fn.([]HttpBefore)
		case []HttpAfter:
			globalHttpAfter = fn.([]HttpAfter)
		}
	}

	if g == nil {
		panic("Group function is nil")
	}

	globalHttpPath = path
	g()
	globalHttpPath = ""
	globalHttpBefore = nil
	globalHttpAfter = nil
}

func (h *Http) SetRoute(method string, path string, v ...interface{}) {

	var m = strings.ToUpper(method)

	path = h.FormatPath(globalHttpPath + path)

	if h.Router == nil {
		h.Router = new(tire.Tire)
	}

	var streamFunction StreamFunction
	var httpFunction HttpFunction
	var before []HttpBefore
	var after []HttpAfter

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
		case func(w http.ResponseWriter, r *http.Request):
			httpFunction = fn.(func(w http.ResponseWriter, r *http.Request))
		case func(t *Stream) func() *Error:
			streamFunction = fn.(func(t *Stream) func() *Error)
		case []HttpBefore:
			before = fn.([]HttpBefore)
		case []HttpAfter:
			after = fn.([]HttpAfter)
		}
	}

	if streamFunction == nil && httpFunction == nil {
		println(m, path, "Stream function and Http function are both nil")
		return
	}

	if streamFunction != nil && httpFunction != nil {
		println(m, path, "Stream function and Http function are both exists")
		return
	}

	var hba = &HBA{}

	hba.StreamFunction = streamFunction
	hba.HttpFunction = httpFunction

	hba.Before = append(globalHttpBefore, before...)
	if passBefore {
		hba.Before = nil
	}
	if forceBefore {
		hba.Before = before
	}

	hba.After = append(globalHttpAfter, after...)
	if passAfter {
		hba.After = nil
	}
	if forceAfter {
		hba.After = after
	}

	hba.Method = m
	hba.Route = []byte(path)

	h.Router.Insert(path, hba)
}

func (h *Http) FormatPath(path string) string {

	if h.IgnoreCase {
		path = strings.ToLower(path)
	}

	return path
}

func (h *Http) GetRoute(method string, path string) *tire.Tire {

	var m = strings.ToUpper(method)

	path = h.FormatPath(path)

	var pathB = []byte(path)

	if h.Router == nil {
		return nil
	}

	var t = h.Router.GetValue(pathB)

	if t == nil {
		return nil
	}

	var hba = t.Data.(*HBA)

	if hba.Method != m {
		return nil
	}

	hba.Path = pathB

	return t

}

func (h *Http) handle(w http.ResponseWriter, r *http.Request) {

	// Get the router
	node := h.GetRoute(r.Method, r.URL.Path)
	if node == nil {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write(nil)
		return
	}

	var hba = node.Data.(*HBA)

	// Get the middleware
	var params = new(Params)
	params.Keys = node.Keys
	params.Values = node.ParseParams(hba.Path)

	var tool = Stream{w, r, nil, params, nil, nil, nil}

	for _, before := range hba.Before {
		context, err := before(&tool)
		if err != nil {
			if h.OnError != nil {
				h.OnError(err)
			}
			return
		}
		tool.Context = context
	}

	if hba.StreamFunction != nil {
		err := hba.StreamFunction(&tool)
		if err != nil {
			if h.OnError != nil {
				h.OnError(err)
			}
			return
		}
	} else {
		hba.HttpFunction(tool.Response, tool.Request)
	}

	for _, after := range hba.After {
		err := after(&tool)
		if err != nil {
			if h.OnError != nil {
				h.OnError(err)
			}
			return
		}
	}
}

func (h *Http) Get(path string, v ...interface{}) {
	h.SetRoute("GET", path, v...)
}

func (h *Http) Post(path string, v ...interface{}) {
	h.SetRoute("POST", path, v...)
}

func (h *Http) Delete(path string, v ...interface{}) {
	h.SetRoute("DELETE", path, v...)
}

func (h *Http) Put(path string, v ...interface{}) {
	h.SetRoute("PUT", path, v...)
}

func (h *Http) Patch(path string, v ...interface{}) {
	h.SetRoute("PATCH", path, v...)
}

func (h *Http) Option(path string, v ...interface{}) {
	h.SetRoute("OPTION", path, v...)
}
