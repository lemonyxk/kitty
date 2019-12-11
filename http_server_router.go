/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-11-25 11:29
**/

package lemo

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Lemo-yxk/tire"

	"github.com/Lemo-yxk/lemo/caller"
	"github.com/Lemo-yxk/lemo/exception"
)

type HttpServerGroupFunction func(handler *HttpServerRouteHandler)

type HttpServerFunction func(t *Stream) func() *exception.Error

type HttpServerBefore func(t *Stream) (Context, func() *exception.Error)

type HttpServerAfter func(t *Stream) func() *exception.Error

var httpServerGlobalBefore []HttpServerBefore
var httpServerGlobalAfter []HttpServerAfter

func SetHttpGlobalBefore(before ...HttpServerBefore) {
	httpServerGlobalBefore = append(httpServerGlobalBefore, before...)
}

func SetHttpGlobalAfter(after ...HttpServerAfter) {
	httpServerGlobalAfter = append(httpServerGlobalAfter, after...)
}

type HttpServerGroup struct {
	path   string
	before []HttpServerBefore
	after  []HttpServerAfter
	router *HttpServerRouter
}

func (group *HttpServerGroup) Route(path string) *HttpServerGroup {
	group.path = path
	return group
}

func (group *HttpServerGroup) Before(before ...HttpServerBefore) *HttpServerGroup {
	group.before = append(group.before, before...)
	return group
}

func (group *HttpServerGroup) After(after ...HttpServerAfter) *HttpServerGroup {
	group.after = append(group.after, after...)
	return group
}

func (group *HttpServerGroup) Handler(fn HttpServerGroupFunction) {

	if group.path == "" {
		panic("group path can not empty")
	}

	fn(&HttpServerRouteHandler{group: group})
}

type HttpServerRouteHandler struct {
	group *HttpServerGroup
}

func (handler *HttpServerRouteHandler) Route(method string, path string) *HttpServerRoute {
	return &HttpServerRoute{path: path, method: method, group: handler.group}
}

func (handler *HttpServerRouteHandler) Get(path string) *HttpServerRoute {
	return handler.Route("GET", path)
}

func (handler *HttpServerRouteHandler) Post(path string) *HttpServerRoute {
	return handler.Route("POST", path)
}

func (handler *HttpServerRouteHandler) Delete(path string) *HttpServerRoute {
	return handler.Route("DELETE", path)
}

func (handler *HttpServerRouteHandler) Put(path string) *HttpServerRoute {
	return handler.Route("PUT", path)
}

func (handler *HttpServerRouteHandler) Patch(path string) *HttpServerRoute {
	return handler.Route("PATCH", path)
}

func (handler *HttpServerRouteHandler) Option(path string) *HttpServerRoute {
	return handler.Route("OPTION", path)
}

type HttpServerRoute struct {
	path        string
	method      string
	before      []HttpServerBefore
	after       []HttpServerAfter
	passBefore  bool
	forceBefore bool
	passAfter   bool
	forceAfter  bool
	group       *HttpServerGroup
}

func (route *HttpServerRoute) Before(before ...HttpServerBefore) *HttpServerRoute {
	route.before = append(route.before, before...)
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

func (route *HttpServerRoute) After(after ...HttpServerAfter) *HttpServerRoute {
	route.after = append(route.after, after...)
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

	if route.path == "" || route.method == "" {
		panic("route path or method can not empty")
	}

	file, line := caller.Caller(1)

	var group = route.group

	var router = route.group.router

	var method = strings.ToUpper(route.method)

	if group == nil {
		group = new(HttpServerGroup)
	}

	var path = router.formatPath(group.path + route.path)

	if router.tire == nil {
		router.tire = new(tire.Tire)
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

	router.tire.Insert(path, hba)
}

type HttpServerRouter struct {
	IgnoreCase   bool
	tire         *tire.Tire
	prefixPath   string
	staticPath   string
	defaultIndex string
}

func (router *HttpServerRouter) GetAllRouters() []*httpServerNode {
	var res []*httpServerNode
	var tires = router.tire.GetAllValue()
	for i := 0; i < len(tires); i++ {
		res = append(res, tires[i].Data.(*httpServerNode))
	}
	return res
}

func (router *HttpServerRouter) SetDefaultIndex(index string) {
	router.defaultIndex = index
}

func (router *HttpServerRouter) SetStaticPath(prefixPath string, staticPath string) {

	if prefixPath == "" {
		panic("prefixPath can not be empty")
	}

	if staticPath == "" {
		panic("staticPath can not be empty")
	}

	absStaticPath, err := filepath.Abs(staticPath)
	if err != nil {
		panic(err)
	}

	info, err := os.Stat(absStaticPath)
	if err != nil {
		panic(err)
	}

	if !info.IsDir() {
		panic("staticPath is not a dir")
	}

	router.prefixPath = prefixPath
	router.staticPath = absStaticPath
	router.defaultIndex = "index.html"
}

func (router *HttpServerRouter) Group(path string) *HttpServerGroup {

	var group = new(HttpServerGroup)

	group.Route(path)

	group.router = router

	return group
}

func (router *HttpServerRouter) getRoute(method string, path string) (*tire.Tire, []byte) {

	if router.tire == nil {
		return nil, nil
	}

	method = strings.ToUpper(method)

	path = router.formatPath(path)

	var pathB = []byte(path)

	var t = router.tire.GetValue(pathB)

	if t == nil {
		return nil, nil
	}

	if t.Data.(*httpServerNode).Method != method {
		return nil, nil
	}

	return t, pathB
}

func (router *HttpServerRouter) formatPath(path string) string {
	if router.IgnoreCase {
		path = strings.ToLower(path)
	}
	return path
}

type httpServerNode struct {
	Info               string
	Route              []byte
	Method             string
	HttpServerFunction HttpServerFunction
	Before             []HttpServerBefore
	After              []HttpServerAfter
}
