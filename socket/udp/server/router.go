/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2021-02-12 22:34
**/

package server

import (
	"strconv"
	"strings"

	"github.com/lemoyxk/structure/tire"

	"github.com/lemoyxk/kitty"
	"github.com/lemoyxk/kitty/socket"
)

type groupFunction func(handler *RouteHandler)

type function func(conn *Conn, stream *socket.Stream) error

type Before func(conn *Conn, stream *socket.Stream) error

type After func(conn *Conn, stream *socket.Stream) error

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
	g.router.tire.Delete(g.path + path)
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
	rh.group.router.tire.Delete(rh.group.path + path)
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

	file, line := kitty.Caller(1)

	var router = r.group.router
	var g = r.group

	if g == nil {
		g = new(group)
	}

	var path = router.formatPath(g.path + r.path)

	if router.tire == nil {
		router.tire = new(tire.Tire)
	}

	var wba = &node{}

	wba.Info = file + ":" + strconv.Itoa(line)

	wba.Function = fn

	wba.Before = append(g.before, r.before...)
	if r.passBefore {
		wba.Before = nil
	}
	if r.forceBefore {
		wba.Before = r.before
	}

	wba.After = append(g.after, r.after...)
	if r.passAfter {
		wba.After = nil
	}
	if r.forceAfter {
		wba.After = r.after
	}

	wba.Before = append(wba.Before, router.globalBefore...)
	wba.After = append(wba.After, router.globalAfter...)

	wba.Route = []byte(path)

	router.tire.Insert(path, wba)

}

type Router struct {
	tire         *tire.Tire
	IgnoreCase   bool
	globalAfter  []After
	globalBefore []Before
}

func (r *Router) SetGlobalBefore(before ...Before) {
	r.globalBefore = append(r.globalBefore, before...)
}

func (r *Router) SetGlobalAfter(after ...After) {
	r.globalAfter = append(r.globalAfter, after...)
}

func (r *Router) Group(path ...string) *group {

	var group = new(group)

	group.path = strings.Join(path, "")

	group.router = r

	return group
}

func (r *Router) Remove(path ...string) {
	r.tire.Delete(strings.Join(path, ""))
}

func (r *Router) Route(path string) *route {
	return (&RouteHandler{group: r.Group("")}).Route(path)
}

func (r *Router) GetAllRouters() []*node {
	var res []*node
	var tires = r.tire.GetAllValue()
	for i := 0; i < len(tires); i++ {
		res = append(res, tires[i].Data.(*node))
	}
	return res
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
