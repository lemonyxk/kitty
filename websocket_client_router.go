package lemo

import (
	"strconv"
	"strings"

	"github.com/Lemo-yxk/lemo/caller"
	"github.com/Lemo-yxk/lemo/exception"

	"github.com/Lemo-yxk/tire"
)

type WebSocketClientGroupFunction func(handler *WebSocketClientRouteHandler)

type WebSocketClientFunction func(c *WebSocketClient, receive *Receive) func() *exception.Error

type WebSocketClientBefore func(c *WebSocketClient, receive *Receive) (Context, func() *exception.Error)

type WebSocketClientAfter func(c *WebSocketClient, receive *Receive) func() *exception.Error

var webSocketClientGlobalBefore []WebSocketClientBefore
var webSocketClientGlobalAfter []WebSocketClientAfter

func SetWebSocketClientBefore(before ...WebSocketClientBefore) {
	webSocketClientGlobalBefore = append(webSocketClientGlobalBefore, before...)
}

func SetWebSocketClientAfter(after ...WebSocketClientAfter) {
	webSocketClientGlobalAfter = append(webSocketClientGlobalAfter, after...)
}

type WebSocketClientGroup struct {
	path   string
	before []WebSocketClientBefore
	after  []WebSocketClientAfter
	router *WebSocketClientRouter
}

func (group *WebSocketClientGroup) Route(path string) *WebSocketClientGroup {
	group.path = path
	return group
}

func (group *WebSocketClientGroup) Before(before ...WebSocketClientBefore) *WebSocketClientGroup {
	group.before = append(group.before, before...)
	return group
}

func (group *WebSocketClientGroup) After(after ...WebSocketClientAfter) *WebSocketClientGroup {
	group.after = append(group.after, after...)
	return group
}

func (group *WebSocketClientGroup) Handler(fn WebSocketClientGroupFunction) {
	if group.path == "" {
		panic("group path can not empty")
	}

	fn(&WebSocketClientRouteHandler{group: group})
}

type WebSocketClientRouteHandler struct {
	group *WebSocketClientGroup
}

func (handler *WebSocketClientRouteHandler) Route(path string) *WebSocketClientRoute {
	return &WebSocketClientRoute{path: path, group: handler.group}
}

type WebSocketClientRoute struct {
	path        string
	before      []WebSocketClientBefore
	after       []WebSocketClientAfter
	socket      *WebSocketClient
	passBefore  bool
	forceBefore bool
	passAfter   bool
	forceAfter  bool
	group       *WebSocketClientGroup
}

func (route *WebSocketClientRoute) Before(before ...WebSocketClientBefore) *WebSocketClientRoute {
	route.before = append(route.before, before...)
	return route
}

func (route *WebSocketClientRoute) PassBefore() *WebSocketClientRoute {
	route.passBefore = true
	return route
}

func (route *WebSocketClientRoute) ForceBefore() *WebSocketClientRoute {
	route.forceBefore = true
	return route
}

func (route *WebSocketClientRoute) After(after ...WebSocketClientAfter) *WebSocketClientRoute {
	route.after = append(route.after, after...)
	return route
}

func (route *WebSocketClientRoute) PassAfter() *WebSocketClientRoute {
	route.passAfter = true
	return route
}

func (route *WebSocketClientRoute) ForceAfter() *WebSocketClientRoute {
	route.forceAfter = true
	return route
}

func (route *WebSocketClientRoute) Handler(fn WebSocketClientFunction) {

	if route.path == "" {
		panic("route path can not empty")
	}

	file, line := caller.RuntimeCaller(1)

	var router = route.group.router
	var group = route.group

	if group == nil {
		group = new(WebSocketClientGroup)
	}

	var path = router.formatPath(group.path + route.path)

	if router.tire == nil {
		router.tire = new(tire.Tire)
	}

	var wba = &WebSocketClientNode{}

	wba.Info = file + ":" + strconv.Itoa(line)

	wba.WebSocketClientFunction = fn

	wba.Before = append(group.before, route.before...)
	if route.passBefore {
		wba.Before = nil
	}
	if route.forceBefore {
		wba.Before = route.before
	}

	wba.After = append(group.after, route.after...)
	if route.passAfter {
		wba.After = nil
	}
	if route.forceAfter {
		wba.After = route.after
	}

	wba.Before = append(wba.Before, webSocketClientGlobalBefore...)
	wba.After = append(wba.After, webSocketClientGlobalAfter...)

	wba.Route = []byte(path)

	router.tire.Insert(path, wba)

}

type WebSocketClientRouter struct {
	tire       *tire.Tire
	IgnoreCase bool
}

func (router *WebSocketClientRouter) GetAllRouters() []*WebSocketClientNode {
	var res []*WebSocketClientNode
	var tires = router.tire.GetAllValue()
	for i := 0; i < len(tires); i++ {
		res = append(res, tires[i].Data.(*WebSocketClientNode))
	}
	return res
}

func (router *WebSocketClientRouter) Group(path string) *WebSocketClientGroup {

	var group = new(WebSocketClientGroup)

	group.Route(path)

	group.router = router

	return group
}

func (router *WebSocketClientRouter) getRoute(path string) (*tire.Tire, []byte) {

	if router.tire == nil {
		return nil, nil
	}

	path = router.formatPath(path)

	var pathB = []byte(path)

	var t = router.tire.GetValue(pathB)

	if t == nil {
		return nil, nil
	}

	return t, pathB
}

func (router *WebSocketClientRouter) formatPath(path string) string {
	if router.IgnoreCase {
		path = strings.ToLower(path)
	}
	return path
}

type WebSocketClientNode struct {
	Info                    string
	Route                   []byte
	WebSocketClientFunction WebSocketClientFunction
	Before                  []WebSocketClientBefore
	After                   []WebSocketClientAfter
}
