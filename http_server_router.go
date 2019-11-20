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
	"github.com/Lemo-yxk/lemo/exception"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/Lemo-yxk/tire"
)

type HttpServerGroupFunction func(this *HttpServer)

type HttpServerFunction func(t *Stream) func() *exception.Error

type HttpServerBefore func(t *Stream) (Context, func() *exception.Error)

type HttpServerAfter func(t *Stream) func() *exception.Error

type ErrorFunction func(err func() *exception.Error)

var httpServerGlobalBefore []HttpServerBefore
var httpServerGlobalAfter []HttpServerAfter

func SetHttpGlobalBefore(before ...HttpServerBefore) {
	httpServerGlobalBefore = append(httpServerGlobalBefore, before...)
}

func SetHttpGlobalAfter(after ...HttpServerAfter) {
	httpServerGlobalAfter = append(httpServerGlobalAfter, after...)
}

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

func (group *httpServerGroup) Before(before ...HttpServerBefore) *httpServerGroup {
	group.before = append(group.before, before...)
	return group
}

func (group *httpServerGroup) After(after ...HttpServerAfter) *httpServerGroup {
	group.after = append(group.after, after...)
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

func (route *httpServerRoute) Before(before ...HttpServerBefore) *httpServerRoute {
	route.before = append(route.before, before...)
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

func (route *httpServerRoute) After(after ...HttpServerAfter) *httpServerRoute {
	route.after = append(route.after, after...)
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

	_, file, line, _ := runtime.Caller(1)

	var h = route.http
	var group = h.group

	var method = strings.ToUpper(route.method)

	if group == nil {
		group = new(httpServerGroup)
	}

	var path = h.formatPath(group.path + route.path)

	if h.tire == nil {
		h.tire = new(tire.Tire)
	}

	var hba = &httpServerNode{}

	hba.Info = file + ":" + strconv.Itoa(line)

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

	hba.Before = append(hba.Before, httpServerGlobalBefore...)
	hba.After = append(hba.After, httpServerGlobalAfter...)

	hba.Method = method

	hba.Route = []byte(path)

	h.tire.Insert(path, hba)

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

func (h *HttpServer) getRoute(method string, path string) (*tire.Tire, []byte) {

	if h.tire == nil {
		return nil, nil
	}

	method = strings.ToUpper(method)

	path = h.formatPath(path)

	var pathB = []byte(path)

	var t = h.tire.GetValue(pathB)

	if t == nil {
		return nil, nil
	}

	if t.Data.(*httpServerNode).Method != method {
		return nil, nil
	}

	return t, pathB
}

func (h *HttpServer) router(w http.ResponseWriter, r *http.Request) {

	// static file
	if h.staticPath != "" && r.Method == http.MethodGet {
		err := h.staticHandler(w, r)
		if err == nil {
			return
		}
	}

	// Get the router
	node, formatPath := h.getRoute(r.Method, r.URL.Path)
	if node == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var nodeData = node.Data.(*httpServerNode)

	// Get the middleware
	var params = new(Params)
	params.Keys = node.Keys
	params.Values = node.ParseParams(formatPath)

	var tool = Stream{w, r, nil, params, nil, nil, nil}

	for i := 0; i < len(nodeData.Before); i++ {
		context, err := nodeData.Before[i](&tool)
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

	for i := 0; i < len(nodeData.After); i++ {
		err := nodeData.After[i](&tool)
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

func (h *HttpServer) Get(path string) *httpServerRoute {
	return h.Route("GET", path)
}

func (h *HttpServer) Post(path string) *httpServerRoute {
	return h.Route("POST", path)
}

func (h *HttpServer) Delete(path string) *httpServerRoute {
	return h.Route("DELETE", path)
}

func (h *HttpServer) Put(path string) *httpServerRoute {
	return h.Route("PUT", path)
}

func (h *HttpServer) Patch(path string) *httpServerRoute {
	return h.Route("PATCH", path)
}

func (h *HttpServer) Option(path string) *httpServerRoute {
	return h.Route("OPTION", path)
}

type httpServerNode struct {
	Info               string
	Route              []byte
	Method             string
	HttpServerFunction HttpServerFunction
	Before             []HttpServerBefore
	After              []HttpServerAfter
}
