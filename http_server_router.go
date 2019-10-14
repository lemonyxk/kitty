/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-10-14 18:46
**/

package lemo

import (
	"net/http"
	"strings"

	"github.com/Lemo-yxk/tire"
)

type HttpServerGroupFunction func(this *HttpServer)

type HttpServerFunction func(t *Stream) func() *Error

type HttpServerBefore func(t *Stream) (Context, func() *Error)

type HttpServerAfter func(t *Stream) func() *Error

type ErrorFunction func(func() *Error)

type httpServerGroup struct {
	path   string
	before []HttpServerBefore
	after  []HttpServerAfter
	http   *HttpServer
}

func (group *httpServerGroup) Route(path string) *httpServerGroup {
	group.path = path
	return group
}

func (group *httpServerGroup) Before(before []HttpServerBefore) *httpServerGroup {
	group.before = before
	return group
}

func (group *httpServerGroup) After(after []HttpServerAfter) *httpServerGroup {
	group.after = after
	return group
}

func (group *httpServerGroup) Handler(fn HttpServerGroupFunction) {
	fn(group.http)
	group.http.group = nil
}

type httpServerRoute struct {
	path        string
	method      string
	before      []HttpServerBefore
	after       []HttpServerAfter
	http        *HttpServer
	passBefore  bool
	forceBefore bool
	passAfter   bool
	forceAfter  bool
}

func (route *httpServerRoute) Route(method string, path string) *httpServerRoute {
	route.path = path
	route.method = method
	return route
}

func (route *httpServerRoute) Before(before []HttpServerBefore) *httpServerRoute {
	route.before = before
	return route
}

func (route *httpServerRoute) PassBefore() *httpServerRoute {
	route.passBefore = true
	return route
}

func (route *httpServerRoute) ForceBefore() *httpServerRoute {
	route.forceBefore = true
	return route
}

func (route *httpServerRoute) After(after []HttpServerAfter) *httpServerRoute {
	route.after = after
	return route
}

func (route *httpServerRoute) PassAfter() *httpServerRoute {
	route.passAfter = true
	return route
}

func (route *httpServerRoute) ForceAfter() *httpServerRoute {
	route.forceAfter = true
	return route
}

func (route *httpServerRoute) Handler(fn HttpServerFunction) {

	var h = route.http
	var group = h.group

	var method = strings.ToUpper(route.method)

	if group == nil {
		group = new(httpServerGroup)
	}

	var path = h.formatPath(group.path + route.path)

	if h.Router == nil {
		h.Router = new(tire.Tire)
	}

	var hba = &httpServerNode{}

	hba.HttpServerFunction = fn

	hba.Before = append(group.before, route.before...)
	if route.passBefore {
		hba.Before = nil
	}
	if route.forceBefore {
		hba.Before = route.before
	}

	hba.After = append(group.after, route.after...)
	if route.passAfter {
		hba.After = nil
	}
	if route.forceAfter {
		hba.After = route.after
	}

	hba.Method = method

	hba.Route = []byte(path)

	h.Router.Insert(path, hba)

	route.http.route = nil
}

func (h *HttpServer) Group(path string) *httpServerGroup {

	var group = new(httpServerGroup)

	group.Route(path)

	group.http = h

	h.group = group

	return group
}

func (h *HttpServer) Route(method string, path string) *httpServerRoute {

	var route = new(httpServerRoute)

	route.Route(method, path)

	route.http = h

	h.route = route

	return route
}

func (h *HttpServer) getRoute(method string, path string) *tire.Tire {

	method = strings.ToUpper(method)

	path = h.formatPath(path)

	var pathB = []byte(path)

	if h.Router == nil {
		return nil
	}

	var t = h.Router.GetValue(pathB)

	if t == nil {
		return nil
	}

	var nodeData = t.Data.(*httpServerNode)

	if nodeData.Method != method {
		return nil
	}

	nodeData.Path = pathB

	return t
}

func (h *HttpServer) router(w http.ResponseWriter, r *http.Request) {

	// Get the router
	node := h.getRoute(r.Method, r.URL.Path)
	if node == nil {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write(nil)
		return
	}

	var nodeData = node.Data.(*httpServerNode)

	// Get the middleware
	var params = new(Params)
	params.Keys = node.Keys
	params.Values = node.ParseParams(nodeData.Path)

	var tool = Stream{w, r, nil, params, nil, nil, nil}

	for _, before := range nodeData.Before {
		context, err := before(&tool)
		if err != nil {
			if h.OnError != nil {
				h.OnError(err)
			}
			return
		}
		tool.Context = context
	}

	if nodeData.HttpServerFunction != nil {
		err := nodeData.HttpServerFunction(&tool)
		if err != nil {
			if h.OnError != nil {
				h.OnError(err)
			}
			return
		}
	}

	for _, after := range nodeData.After {
		err := after(&tool)
		if err != nil {
			if h.OnError != nil {
				h.OnError(err)
			}
			return
		}
	}
}

func (h *HttpServer) formatPath(path string) string {
	if h.IgnoreCase {
		path = strings.ToLower(path)
	}
	return path
}

func (h *HttpServer) Get(path string, fn HttpServerFunction) {
	h.Route("GET", path).Handler(fn)
}

func (h *HttpServer) Post(path string, fn HttpServerFunction) {
	h.Route("POST", path).Handler(fn)
}

func (h *HttpServer) Delete(path string, fn HttpServerFunction) {
	h.Route("DELETE", path).Handler(fn)
}

func (h *HttpServer) Put(path string, fn HttpServerFunction) {
	h.Route("PUT", path).Handler(fn)
}

func (h *HttpServer) Patch(path string, fn HttpServerFunction) {
	h.Route("PATCH", path).Handler(fn)
}

func (h *HttpServer) Option(path string, fn HttpServerFunction) {
	h.Route("OPTION", path).Handler(fn)
}

type httpServerNode struct {
	Path               []byte
	Route              []byte
	Method             string
	HttpServerFunction HttpServerFunction
	Before             []HttpServerBefore
	After              []HttpServerAfter
}
