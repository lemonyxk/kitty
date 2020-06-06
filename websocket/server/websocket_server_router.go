package server

import (
	"strconv"
	"strings"

	"github.com/Lemo-yxk/lemo"
	"github.com/Lemo-yxk/lemo/caller"
	"github.com/Lemo-yxk/lemo/container/tire"
	"github.com/Lemo-yxk/lemo/exception"
)

type groupFunction func(handler *RouteHandler)

type function func(conn *WebSocket, receive *lemo.Receive) exception.Error

type before func(conn *WebSocket, receive *lemo.Receive) (lemo.Context, exception.Error)

type after func(conn *WebSocket, receive *lemo.Receive) exception.Error

var globalBefore []before
var globalAfter []after

func SetBefore(before ...before) {
	globalBefore = append(globalBefore, before...)
}

func SetAfter(after ...after) {
	globalAfter = append(globalAfter, after...)
}

type group struct {
	path   string
	before []before
	after  []after
	router *Router
}

func (group *group) Route(path string) *group {
	group.path = path
	return group
}

func (group *group) Before(before ...before) *group {
	group.before = append(group.before, before...)
	return group
}

func (group *group) After(after ...after) *group {
	group.after = append(group.after, after...)
	return group
}

func (group *group) Handler(fn groupFunction) {
	if group.path == "" {
		panic("group path can not empty")
	}

	fn(&RouteHandler{group: group})
}

type RouteHandler struct {
	group *group
}

func (handler *RouteHandler) Route(path string) *route {
	return &route{path: path, group: handler.group}
}

type route struct {
	path        string
	before      []before
	after       []after
	socket      *Server
	passBefore  bool
	forceBefore bool
	passAfter   bool
	forceAfter  bool
	group       *group
}

func (route *route) Before(before ...before) *route {
	route.before = append(route.before, before...)
	return route
}

func (route *route) PassBefore() *route {
	route.passBefore = true
	return route
}

func (route *route) ForceBefore() *route {
	route.forceBefore = true
	return route
}

func (route *route) After(after ...after) *route {
	route.after = append(route.after, after...)
	return route
}

func (route *route) PassAfter() *route {
	route.passAfter = true
	return route
}

func (route *route) ForceAfter() *route {
	route.forceAfter = true
	return route
}

func (route *route) Handler(fn function) {

	if route.path == "" {
		panic("route path can not empty")
	}

	file, line := caller.Caller(1)

	var router = route.group.router
	var g = route.group

	if g == nil {
		g = new(group)
	}

	var path = router.formatPath(g.path + route.path)

	if router.tire == nil {
		router.tire = new(tire.Tire)
	}

	var wba = &node{}

	wba.Info = file + ":" + strconv.Itoa(line)

	wba.Function = fn

	wba.Before = append(g.before, route.before...)
	if route.passBefore {
		wba.Before = nil
	}
	if route.forceBefore {
		wba.Before = route.before
	}

	wba.After = append(g.after, route.after...)
	if route.passAfter {
		wba.After = nil
	}
	if route.forceAfter {
		wba.After = route.after
	}

	wba.Before = append(wba.Before, globalBefore...)
	wba.After = append(wba.After, globalAfter...)

	wba.Route = []byte(path)

	router.tire.Insert(path, wba)

}

type Router struct {
	tire       *tire.Tire
	IgnoreCase bool
}

func (router *Router) Group(path string) *group {

	var group = new(group)

	group.Route(path)

	group.router = router

	return group
}

func (router *Router) Route(path string) *route {
	return (&RouteHandler{group: router.Group("")}).Route(path)
}

func (router *Router) GetAllRouters() []*node {
	var res []*node
	var tires = router.tire.GetAllValue()
	for i := 0; i < len(tires); i++ {
		res = append(res, tires[i].Data.(*node))
	}
	return res
}

func (router *Router) getRoute(path string) (*tire.Tire, []byte) {

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

func (router *Router) formatPath(path string) string {
	if router.IgnoreCase {
		path = strings.ToLower(path)
	}
	return path
}

type node struct {
	Info     string
	Route    []byte
	Function function
	Before   []before
	After    []after
}
