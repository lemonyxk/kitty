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

type HttpServer struct {
	IgnoreCase bool
	Router     *tire.Tire
	OnError    ErrorFunction

	group *HttpServerGroup
	route *HttpServerRoute
}

type HttpServerGroup struct {
	path   string
	before []HttpServerBefore
	after  []HttpServerAfter
	http   *HttpServer
}

func (group *HttpServerGroup) Route(path string) *HttpServerGroup {
	group.path = path
	return group
}

func (group *HttpServerGroup) Before(before []HttpServerBefore) *HttpServerGroup {
	group.before = before
	return group
}

func (group *HttpServerGroup) After(after []HttpServerAfter) *HttpServerGroup {
	group.after = after
	return group
}

func (group *HttpServerGroup) Handler(fn HttpServerGroupFunction) {
	fn(group.http)
	group.http.group = nil
}

type HttpServerRoute struct {
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

func (route *HttpServerRoute) Route(method string, path string) *HttpServerRoute {
	route.path = path
	route.method = method
	return route
}

func (route *HttpServerRoute) Before(before []HttpServerBefore) *HttpServerRoute {
	route.before = before
	return route
}

func (route *HttpServerRoute) PassBefore() *HttpServerRoute {
	route.passBefore = true
	return route
}

func (route *HttpServerRoute) ForceBefore() *HttpServerRoute {
	route.forceBefore = true
	return route
}

func (route *HttpServerRoute) After(after []HttpServerAfter) *HttpServerRoute {
	route.after = after
	return route
}

func (route *HttpServerRoute) PassAfter() *HttpServerRoute {
	route.passAfter = true
	return route
}

func (route *HttpServerRoute) ForceAfter() *HttpServerRoute {
	route.forceAfter = true
	return route
}

func (route *HttpServerRoute) Handler(fn HttpServerFunction) {

	var h = route.http
	var group = h.group

	var method = strings.ToUpper(route.method)

	if group == nil {
		group = new(HttpServerGroup)
	}

	var path = h.FormatPath(group.path + route.path)

	if h.Router == nil {
		h.Router = new(tire.Tire)
	}

	var hba = &HttpServerNode{}

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

func (h *HttpServer) Group(path string) *HttpServerGroup {

	var group = new(HttpServerGroup)

	group.Route(path)

	group.http = h

	h.group = group

	return group
}

func (h *HttpServer) Route(method string, path string) *HttpServerRoute {

	var route = new(HttpServerRoute)

	route.Route(method, path)

	route.http = h

	h.route = route

	return route
}

func (h *HttpServer) getRoute(method string, path string) *tire.Tire {

	method = strings.ToUpper(method)

	path = h.FormatPath(path)

	var pathB = []byte(path)

	if h.Router == nil {
		return nil
	}

	var t = h.Router.GetValue(pathB)

	if t == nil {
		return nil
	}

	var nodeData = t.Data.(*HttpServerNode)

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

	var nodeData = node.Data.(*HttpServerNode)

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

func (h *HttpServer) FormatPath(path string) string {
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
func (h *HttpServer) Ready() {

}

type HttpServerNode struct {
	Path               []byte
	Route              []byte
	Method             string
	HttpServerFunction HttpServerFunction
	Before             []HttpServerBefore
	After              []HttpServerAfter
}
