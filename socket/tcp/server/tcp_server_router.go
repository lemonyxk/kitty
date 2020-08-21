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

func (g *group) Route(path string) *group {
	g.path = path
	return g
}

func (g *group) Before(before ...Before) *group {
	g.before = append(g.before, before...)
	return g
}

func (g *group) After(after ...After) *group {
	g.after = append(g.after, after...)
	return g
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

	var sba = &node{}

	sba.Info = file + ":" + strconv.Itoa(line)

	sba.Function = fn

	sba.Before = append(g.before, r.before...)
	if r.passBefore {
		sba.Before = nil
	}
	if r.forceBefore {
		sba.Before = r.before
	}

	sba.After = append(g.after, r.after...)
	if r.passAfter {
		sba.After = nil
	}
	if r.forceAfter {
		sba.After = r.after
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

	var group = new(group)

	group.Route(strings.Join(path, ""))

	group.router = r

	return group
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
