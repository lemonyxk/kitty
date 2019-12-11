/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-10-16 16:10
**/

package lemo

import (
	"strconv"
	"strings"

	"github.com/Lemo-yxk/tire"

	"github.com/Lemo-yxk/lemo/caller"
	"github.com/Lemo-yxk/lemo/exception"
)

type SocketClientGroupFunction func(handler *SocketClientRouteHandler)

type SocketClientFunction func(c *SocketClient, receive *Receive) exception.ErrorFunc

type SocketClientBefore func(c *SocketClient, receive *Receive) (Context, exception.ErrorFunc)

type SocketClientAfter func(c *SocketClient, receive *Receive) exception.ErrorFunc

var socketClientGlobalBefore []SocketClientBefore
var socketClientGlobalAfter []SocketClientAfter

func SetSocketClientBefore(before ...SocketClientBefore) {
	socketClientGlobalBefore = append(socketClientGlobalBefore, before...)
}

func SetSocketClientAfter(after ...SocketClientAfter) {
	socketClientGlobalAfter = append(socketClientGlobalAfter, after...)
}

type SocketClientGroup struct {
	path   string
	before []SocketClientBefore
	after  []SocketClientAfter
	router *SocketClientRouter
}

func (group *SocketClientGroup) Route(path string) *SocketClientGroup {
	group.path = path
	return group
}

func (group *SocketClientGroup) Before(before ...SocketClientBefore) *SocketClientGroup {
	group.before = append(group.before, before...)
	return group
}

func (group *SocketClientGroup) After(after ...SocketClientAfter) *SocketClientGroup {
	group.after = append(group.after, after...)
	return group
}

func (group *SocketClientGroup) Handler(fn SocketClientGroupFunction) {
	fn(&SocketClientRouteHandler{group: group})
}

type SocketClientRouteHandler struct {
	group *SocketClientGroup
}

func (handler *SocketClientRouteHandler) Route(path string) *SocketClientRoute {
	return &SocketClientRoute{path: path, group: handler.group}
}

type SocketClientRoute struct {
	path        string
	before      []SocketClientBefore
	after       []SocketClientAfter
	passBefore  bool
	forceBefore bool
	passAfter   bool
	forceAfter  bool
	group       *SocketClientGroup
}

func (route *SocketClientRoute) Before(before ...SocketClientBefore) *SocketClientRoute {
	route.before = append(route.before, before...)
	return route
}

func (route *SocketClientRoute) PassBefore() *SocketClientRoute {
	route.passBefore = true
	return route
}

func (route *SocketClientRoute) ForceBefore() *SocketClientRoute {
	route.forceBefore = true
	return route
}

func (route *SocketClientRoute) After(after ...SocketClientAfter) *SocketClientRoute {
	route.after = append(route.after, after...)
	return route
}

func (route *SocketClientRoute) PassAfter() *SocketClientRoute {
	route.passAfter = true
	return route
}

func (route *SocketClientRoute) ForceAfter() *SocketClientRoute {
	route.forceAfter = true
	return route
}

func (route *SocketClientRoute) Handler(fn SocketClientFunction) {

	if route.path == "" {
		panic("route path can not empty")
	}

	file, line := caller.Caller(1)

	var router = route.group.router
	var group = route.group

	if group == nil {
		group = new(SocketClientGroup)
	}

	var path = router.formatPath(group.path + route.path)

	if router.tire == nil {
		router.tire = new(tire.Tire)
	}

	var cba = &SocketClientNode{}

	cba.Info = file + ":" + strconv.Itoa(line)

	cba.SocketClientFunction = fn

	cba.Before = append(group.before, route.before...)
	if route.passBefore {
		cba.Before = nil
	}
	if route.forceBefore {
		cba.Before = route.before
	}

	cba.After = append(group.after, route.after...)
	if route.passAfter {
		cba.After = nil
	}
	if route.forceAfter {
		cba.After = route.after
	}

	cba.Before = append(cba.Before, socketClientGlobalBefore...)
	cba.After = append(cba.After, socketClientGlobalAfter...)

	cba.Route = []byte(path)

	router.tire.Insert(path, cba)

}

type SocketClientRouter struct {
	IgnoreCase bool
	tire       *tire.Tire
}

func (router *SocketClientRouter) GetAllRouters() []*SocketClientNode {
	var res []*SocketClientNode
	var tires = router.tire.GetAllValue()
	for i := 0; i < len(tires); i++ {
		res = append(res, tires[i].Data.(*SocketClientNode))
	}
	return res
}

func (router *SocketClientRouter) Group(path string) *SocketClientGroup {

	var group = new(SocketClientGroup)

	group.Route(path)

	group.router = router

	return group
}

func (router *SocketClientRouter) getRoute(path string) (*tire.Tire, []byte) {

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

func (router *SocketClientRouter) formatPath(path string) string {
	if router.IgnoreCase {
		path = strings.ToLower(path)
	}
	return path
}

type SocketClientNode struct {
	Info                 string
	Route                []byte
	SocketClientFunction SocketClientFunction
	Before               []SocketClientBefore
	After                []SocketClientAfter
}
