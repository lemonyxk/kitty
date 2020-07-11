/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-10-12 15:52
**/

package server

import (
	"strconv"
	"strings"

	"github.com/lemoyxk/structure/tire"

	"github.com/lemoyxk/lemo"
)

type groupFunction func(handler *RouteHandler)

type function func(conn *Socket, receive *kitty.Receive) error

type Before func(conn *Socket, receive *kitty.Receive) (kitty.Context, error)

type After func(conn *Socket, receive *kitty.Receive) error

type group struct {
	path   string
	before []Before
	after  []After
	router *Router
}

func (group *group) Route(path string) *group {
	group.path = path
	return group
}

func (group *group) Before(before ...Before) *group {
	group.before = append(group.before, before...)
	return group
}

func (group *group) After(after ...After) *group {
	group.after = append(group.after, after...)
	return group
}

func (group *group) Handler(fn groupFunction) {
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
	before      []Before
	after       []After
	socket      *Server
	passBefore  bool
	forceBefore bool
	passAfter   bool
	forceAfter  bool
	group       *group
}

func (route *route) Before(before ...Before) *route {
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

func (route *route) After(after ...After) *route {
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

	file, line := kitty.Caller(1)

	var router = route.group.router
	var g = route.group

	if g == nil {
		g = new(group)
	}

	var path = router.formatPath(g.path + route.path)

	if router.tire == nil {
		router.tire = new(tire.Tire)
	}

	var sba = &node{}

	sba.Info = file + ":" + strconv.Itoa(line)

	sba.Function = fn

	sba.Before = append(g.before, route.before...)
	if route.passBefore {
		sba.Before = nil
	}
	if route.forceBefore {
		sba.Before = route.before
	}

	sba.After = append(g.after, route.after...)
	if route.passAfter {
		sba.After = nil
	}
	if route.forceAfter {
		sba.After = route.after
	}

	sba.Before = append(sba.Before, router.globalBefore...)
	sba.After = append(sba.After, router.globalAfter...)

	sba.Route = []byte(path)

	router.tire.Insert(path, sba)

}

type Router struct {
	tire         *tire.Tire
	IgnoreCase   bool
	globalAfter  []After
	globalBefore []Before
}

func (router *Router) SetGlobalBefore(before ...Before) {
	router.globalBefore = append(router.globalBefore, before...)
}

func (router *Router) SetGlobalAfter(after ...After) {
	router.globalAfter = append(router.globalAfter, after...)
}

func (router *Router) GetAllRouters() []*node {
	var res []*node
	var tires = router.tire.GetAllValue()
	for i := 0; i < len(tires); i++ {
		res = append(res, tires[i].Data.(*node))
	}
	return res
}

func (router *Router) Group(path ...string) *group {

	var group = new(group)

	group.Route(strings.Join(path, ""))

	group.router = router

	return group
}

func (router *Router) Route(path string) *route {
	return (&RouteHandler{group: router.Group("")}).Route(path)
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
	Before   []Before
	After    []After
}
