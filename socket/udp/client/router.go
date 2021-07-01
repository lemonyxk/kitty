/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2021-02-13 17:39
**/

package client

import (
	"strconv"
	"strings"

	"github.com/lemoyxk/caller"
	"github.com/lemoyxk/structure/tire"

	"github.com/lemoyxk/kitty/socket"
)

type groupFunction func(handler *RouteHandler)

type function func(client *Client, stream *socket.Stream) error

type Before func(client *Client, stream *socket.Stream) error

type After func(client *Client, stream *socket.Stream) error

type group struct {
	path   string
	before []Before
	after  []After
	router *Router
}

func (g *group) Before(before ...Before) *group {
	g.before = append(g.before, before...)
	return g
}

func (g *group) After(after ...After) *group {
	g.after = append(g.after, after...)
	return g
}

func (g *group) Remove(path string) {
	if g.router.tire == nil {
		return
	}
	var dp = g.path + path
	if g.router.IgnoreCase {
		dp = strings.ToLower(dp)
	}
	g.router.tire.Delete(dp)
}

func (g *group) Handler(fn groupFunction) {
	fn(&RouteHandler{group: g})
}

type RouteHandler struct {
	group *group
}

func (rh *RouteHandler) Route(path string) *route {
	return &route{path: path, group: rh.group}
}

func (rh *RouteHandler) Remove(path string) {
	if rh.group.router.tire == nil {
		return
	}
	var dp = rh.group.path + path
	if rh.group.router.IgnoreCase {
		dp = strings.ToLower(dp)
	}
	rh.group.router.tire.Delete(dp)
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

func (r *route) Before(before ...Before) *route {
	r.before = append(r.before, before...)
	return r
}

func (r *route) PassBefore() *route {
	r.passBefore = true
	return r
}

func (r *route) ForceBefore() *route {
	r.forceBefore = true
	return r
}

func (r *route) After(after ...After) *route {
	r.after = append(r.after, after...)
	return r
}

func (r *route) PassAfter() *route {
	r.passAfter = true
	return r
}

func (r *route) ForceAfter() *route {
	r.forceAfter = true
	return r
}

func (r *route) Handler(fn function) {

	if r.path == "" {
		panic("route path can not empty")
	}

	file, line := caller.Deep(2)

	var router = r.group.router
	var g = r.group

	if g == nil {
		g = new(group)
	}

	var path = router.formatPath(g.path + r.path)

	if router.tire == nil {
		router.tire = new(tire.Tire)
	}

	var cba = &node{}

	cba.Info = file + ":" + strconv.Itoa(line)

	cba.Function = fn

	cba.Before = append(g.before, r.before...)
	if r.passBefore {
		cba.Before = nil
	}
	if r.forceBefore {
		cba.Before = r.before
	}

	cba.After = append(g.after, r.after...)
	if r.passAfter {
		cba.After = nil
	}
	if r.forceAfter {
		cba.After = r.after
	}

	cba.Before = append(cba.Before, router.globalBefore...)
	cba.After = append(cba.After, router.globalAfter...)

	cba.Route = []byte(path)

	router.tire.Insert(path, cba)

}

type Router struct {
	IgnoreCase   bool
	tire         *tire.Tire
	globalAfter  []After
	globalBefore []Before
}

func (r *Router) SetGlobalBefore(before ...Before) {
	r.globalBefore = append(r.globalBefore, before...)
}

func (r *Router) SetGlobalAfter(after ...After) {
	r.globalAfter = append(r.globalAfter, after...)
}

func (r *Router) GetAllRouters() []*node {
	var res []*node
	var tires = r.tire.GetAllValue()
	for i := 0; i < len(tires); i++ {
		res = append(res, tires[i].Data.(*node))
	}
	return res
}

func (r *Router) Group(path ...string) *group {

	var g = new(group)

	g.path = strings.Join(path, "")

	g.router = r

	return g
}

func (r *Router) Remove(path ...string) {
	if r.tire == nil {
		return
	}
	var dp = strings.Join(path, "")
	if r.IgnoreCase {
		dp = strings.ToLower(dp)
	}
	r.tire.Delete(dp)
}

func (r *Router) Route(path string) *route {
	return (&RouteHandler{group: r.Group("")}).Route(path)
}

func (r *Router) getRoute(path string) (*tire.Tire, []byte) {

	if r.tire == nil {
		return nil, nil
	}

	path = r.formatPath(path)

	var pathB = []byte(path)

	var t = r.tire.GetValue(pathB)

	if t == nil {
		return nil, nil
	}

	return t, pathB
}

func (r *Router) formatPath(path string) string {
	if r.IgnoreCase {
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
