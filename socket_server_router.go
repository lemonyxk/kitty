/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-10-12 15:52
**/

package lemo

import (
	"runtime"
	"strconv"
	"strings"

	"github.com/Lemo-yxk/lemo/exception"

	"github.com/Lemo-yxk/tire"
)

type SocketServerGroupFunction func(route *SocketServerRoute)

type SocketServerFunction func(conn *Socket, receive *Receive) func() *exception.Error

type SocketServerBefore func(conn *Socket, receive *Receive) (Context, func() *exception.Error)

type SocketServerAfter func(conn *Socket, receive *Receive) func() *exception.Error

var socketServerGlobalBefore []SocketServerBefore
var socketServerGlobalAfter []SocketServerAfter

func SetSocketServerBefore(before ...SocketServerBefore) {
	socketServerGlobalBefore = append(socketServerGlobalBefore, before...)
}

func SetSocketServerAfter(after ...SocketServerAfter) {
	socketServerGlobalAfter = append(socketServerGlobalAfter, after...)
}

type SocketServerGroup struct {
	path   string
	before []SocketServerBefore
	after  []SocketServerAfter
	router *SocketServerRouter
}

func (group *SocketServerGroup) Route(path string) *SocketServerGroup {
	group.path = path
	return group
}

func (group *SocketServerGroup) Before(before ...SocketServerBefore) *SocketServerGroup {
	group.before = append(group.before, before...)
	return group
}

func (group *SocketServerGroup) After(after ...SocketServerAfter) *SocketServerGroup {
	group.after = append(group.after, after...)
	return group
}

func (group *SocketServerGroup) Handler(fn SocketServerGroupFunction) {
	if group.path == "" {
		panic("group path can not empty")
	}
	var route = new(SocketServerRoute)
	route.group = group
	fn(route)
}

type SocketServerRoute struct {
	path        string
	before      []SocketServerBefore
	after       []SocketServerAfter
	socket      *SocketServer
	passBefore  bool
	forceBefore bool
	passAfter   bool
	forceAfter  bool
	group       *SocketServerGroup
}

func (route *SocketServerRoute) Route(path string) *SocketServerRoute {
	route.path = path
	return route
}

func (route *SocketServerRoute) Before(before ...SocketServerBefore) *SocketServerRoute {
	route.before = append(route.before, before...)
	return route
}

func (route *SocketServerRoute) PassBefore() *SocketServerRoute {
	route.passBefore = true
	return route
}

func (route *SocketServerRoute) ForceBefore() *SocketServerRoute {
	route.forceBefore = true
	return route
}

func (route *SocketServerRoute) After(after ...SocketServerAfter) *SocketServerRoute {
	route.after = append(route.after, after...)
	return route
}

func (route *SocketServerRoute) PassAfter() *SocketServerRoute {
	route.passAfter = true
	return route
}

func (route *SocketServerRoute) ForceAfter() *SocketServerRoute {
	route.forceAfter = true
	return route
}

func (route *SocketServerRoute) Handler(fn SocketServerFunction) {

	if route.path == "" {
		panic("route path can not empty")
	}

	_, file, line, _ := runtime.Caller(1)

	var router = route.group.router
	var group = route.group

	if group == nil {
		group = new(SocketServerGroup)
	}

	var path = router.formatPath(group.path + route.path)

	if router.tire == nil {
		router.tire = new(tire.Tire)
	}

	var sba = &SocketServerNode{}

	sba.Info = file + ":" + strconv.Itoa(line)

	sba.SocketServerFunction = fn

	sba.Before = append(group.before, route.before...)
	if route.passBefore {
		sba.Before = nil
	}
	if route.forceBefore {
		sba.Before = route.before
	}

	sba.After = append(group.after, route.after...)
	if route.passAfter {
		sba.After = nil
	}
	if route.forceAfter {
		sba.After = route.after
	}

	sba.Before = append(sba.Before, socketServerGlobalBefore...)
	sba.After = append(sba.After, socketServerGlobalAfter...)

	sba.Route = []byte(path)

	router.tire.Insert(path, sba)

}

type SocketServerRouter struct {
	tire       *tire.Tire
	IgnoreCase bool
}

func (router *SocketServerRouter) GetAllRouters() []*SocketServerNode {
	var res []*SocketServerNode
	var tires = router.tire.GetAllValue()
	for i := 0; i < len(router.tire.GetAllValue()); i++ {
		res = append(res, tires[i].Data.(*SocketServerNode))
	}
	return res
}

func (router *SocketServerRouter) Group(path string) *SocketServerGroup {

	var group = new(SocketServerGroup)

	group.Route(path)

	group.router = router

	return group
}

func (router *SocketServerRouter) getRoute(path string) (*tire.Tire, []byte) {

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

func (router *SocketServerRouter) formatPath(path string) string {
	if router.IgnoreCase {
		path = strings.ToLower(path)
	}
	return path
}

type SocketServerNode struct {
	Info                 string
	Route                []byte
	SocketServerFunction SocketServerFunction
	Before               []SocketServerBefore
	After                []SocketServerAfter
}
