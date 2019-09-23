package ws

import (
	"net/http"
	"strings"
)

type GroupFunction func()

type StreamFunction func(t *Stream)

type HttpFunction func(w http.ResponseWriter, r *http.Request)

type Before func(t *Stream) (interface{}, error)

type After func(t *Stream) error

type Http struct {
	IgnoreCase bool
	Router     *tire
}

type Hba struct {
	Method         string
	StreamFunction StreamFunction
	HttpFunction   HttpFunction
	Before         []Before
	After          []After
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
		path = strings.ToLower(path)
	}

	if h.Router == nil {
		h.Router = new(tire)
	}

	var hba = &Hba{}

	var streamFunction StreamFunction
	var httpFunction HttpFunction
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
		case func(w http.ResponseWriter, r *http.Request):
			httpFunction = fn.(func(w http.ResponseWriter, r *http.Request))
		case func(t *Stream):
			streamFunction = fn.(func(t *Stream))
		case []Before:
			before = fn.([]Before)
		case []After:
			after = fn.([]After)
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

	hba.StreamFunction = streamFunction
	hba.HttpFunction = httpFunction

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

	hba.Method = m

	h.Router.Insert(path, hba)
}

func (h *Http) GetRoute(method string, path string) *tire {

	var m = strings.ToUpper(method)

	if h.IgnoreCase {
		path = strings.ToLower(path)
	}

	if h.Router == nil {
		return nil
	}

	var tire = h.Router.GetValue([]byte(path))

	if tire == nil {
		return nil
	}

	var hba = tire.data.(*Hba)

	if hba.Method != m {
		return nil
	}

	return tire

	// handle, p, tsr := h.Router.getValue(path)
	//
	// if handle.Method != m {
	// 	return nil, nil, false
	// }
	//
	// return handle, p, tsr

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
