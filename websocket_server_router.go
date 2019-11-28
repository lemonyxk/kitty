package lemo

import (
	"strconv"
	"strings"

	"github.com/Lemo-yxk/lemo/caller"
	"github.com/Lemo-yxk/lemo/exception"

	"github.com/Lemo-yxk/tire"
)

type WebSocketServerGroupFunction func(handler *WebSocketServerRouteHandler)

type WebSocketServerFunction func(conn *WebSocket, receive *Receive) func() *exception.Error

type WebSocketServerBefore func(conn *WebSocket, receive *Receive) (Context, func() *exception.Error)

type WebSocketServerAfter func(conn *WebSocket, receive *Receive) func() *exception.Error

var webSocketServerGlobalBefore []WebSocketServerBefore
var webSocketServerGlobalAfter []WebSocketServerAfter

func SetWebSocketServerBefore(before ...WebSocketServerBefore) {
	webSocketServerGlobalBefore = append(webSocketServerGlobalBefore, before...)
}

func SetWebSocketServerAfter(after ...WebSocketServerAfter) {
	webSocketServerGlobalAfter = append(webSocketServerGlobalAfter, after...)
}

type WebSocketServerGroup struct {
	path   string
	before []WebSocketServerBefore
	after  []WebSocketServerAfter
	router *WebSocketServerRouter
}

func (group *WebSocketServerGroup) Route(path string) *WebSocketServerGroup {
	group.path = path
	return group
}

func (group *WebSocketServerGroup) Before(before ...WebSocketServerBefore) *WebSocketServerGroup {
	group.before = append(group.before, before...)
	return group
}

func (group *WebSocketServerGroup) After(after ...WebSocketServerAfter) *WebSocketServerGroup {
	group.after = append(group.after, after...)
	return group
}

func (group *WebSocketServerGroup) Handler(fn WebSocketServerGroupFunction) {
	if group.path == "" {
		panic("group path can not empty")
	}

	fn(&WebSocketServerRouteHandler{group: group})
}

type WebSocketServerRouteHandler struct {
	group *WebSocketServerGroup
}

func (handler *WebSocketServerRouteHandler) Route(path string) *WebSocketServerRoute {
	return &WebSocketServerRoute{path: path, group: handler.group}
}

type WebSocketServerRoute struct {
	path        string
	before      []WebSocketServerBefore
	after       []WebSocketServerAfter
	socket      *WebSocketServer
	passBefore  bool
	forceBefore bool
	passAfter   bool
	forceAfter  bool
	group       *WebSocketServerGroup
}

func (route *WebSocketServerRoute) Before(before ...WebSocketServerBefore) *WebSocketServerRoute {
	route.before = append(route.before, before...)
	return route
}

func (route *WebSocketServerRoute) PassBefore() *WebSocketServerRoute {
	route.passBefore = true
	return route
}

func (route *WebSocketServerRoute) ForceBefore() *WebSocketServerRoute {
	route.forceBefore = true
	return route
}

func (route *WebSocketServerRoute) After(after ...WebSocketServerAfter) *WebSocketServerRoute {
	route.after = append(route.after, after...)
	return route
}

func (route *WebSocketServerRoute) PassAfter() *WebSocketServerRoute {
	route.passAfter = true
	return route
}

func (route *WebSocketServerRoute) ForceAfter() *WebSocketServerRoute {
	route.forceAfter = true
	return route
}

func (route *WebSocketServerRoute) Handler(fn WebSocketServerFunction) {

	if route.path == "" {
		panic("route path can not empty")
	}

	file, line := caller.RuntimeCaller(1)

	var router = route.group.router
	var group = route.group

	if group == nil {
		group = new(WebSocketServerGroup)
	}

	var path = router.formatPath(group.path + route.path)

	if router.tire == nil {
		router.tire = new(tire.Tire)
	}

	var wba = &WebSocketServerNode{}

	wba.Info = file + ":" + strconv.Itoa(line)

	wba.WebSocketServerFunction = fn

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

	wba.Before = append(wba.Before, webSocketServerGlobalBefore...)
	wba.After = append(wba.After, webSocketServerGlobalAfter...)

	wba.Route = []byte(path)

	router.tire.Insert(path, wba)

}

type WebSocketServerRouter struct {
	tire       *tire.Tire
	IgnoreCase bool
}

func (router *WebSocketServerRouter) Group(path string) *WebSocketServerGroup {

	var group = new(WebSocketServerGroup)

	group.Route(path)

	group.router = router

	return group
}

func (router *WebSocketServerRouter) GetAllRouters() []*WebSocketServerNode {
	var res []*WebSocketServerNode
	var tires = router.tire.GetAllValue()
	for i := 0; i < len(tires); i++ {
		res = append(res, tires[i].Data.(*WebSocketServerNode))
	}
	return res
}

func (router *WebSocketServerRouter) getRoute(path string) (*tire.Tire, []byte) {

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

func (router *WebSocketServerRouter) formatPath(path string) string {
	if router.IgnoreCase {
		path = strings.ToLower(path)
	}
	return path
}

type WebSocketServerNode struct {
	Info                    string
	Route                   []byte
	WebSocketServerFunction WebSocketServerFunction
	Before                  []WebSocketServerBefore
	After                   []WebSocketServerAfter
}
