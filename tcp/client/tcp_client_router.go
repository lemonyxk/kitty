/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-10-16 16:10
**/

package client

import (
	"strconv"
	"strings"

	"github.com/Lemo-yxk/lemo"
	"github.com/Lemo-yxk/lemo/caller"
	"github.com/Lemo-yxk/lemo/container/tire"
	"github.com/Lemo-yxk/lemo/exception"
	"github.com/Lemo-yxk/lemo/utils"
)

type groupFunction func(handler *RouteHandler)

type function func(c *Client, receive *lemo.Receive) exception.Error

type Before func(c *Client, receive *lemo.Receive) (lemo.Context, exception.Error)

type After func(c *Client, receive *lemo.Receive) exception.Error

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

	var cba = &node{}

	cba.Info = file + ":" + strconv.Itoa(line)

	cba.Function = fn

	cba.Before = append(g.before, route.before...)
	if route.passBefore {
		cba.Before = nil
	}
	if route.forceBefore {
		cba.Before = route.before
	}

	cba.After = append(g.after, route.after...)
	if route.passAfter {
		cba.After = nil
	}
	if route.forceAfter {
		cba.After = route.after
	}

	cba.Before = append(cba.Before, router.globalBefore...)
	cba.After = append(cba.After, router.globalAfter...)

	cba.Route = utils.Conv.StringToBytes(path)

	router.tire.Insert(path, cba)

}

type Router struct {
	IgnoreCase   bool
	tire         *tire.Tire
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

	var pathB = utils.Conv.StringToBytes(path)

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
