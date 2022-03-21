package server

import (
	"strconv"
	"strings"

	"github.com/lemonyxk/caller"
	"github.com/lemonyxk/structure/v3/tire"

	"github.com/lemonyxk/kitty/v2/socket"
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

func (g *group) Remove(path ...string) {
	if g.router.tire == nil {
		return
	}
	for i := 0; i < len(path); i++ {
		var dp = g.path + path[0]
		if !g.router.StrictMode {
			dp = strings.ToLower(dp)
		}
		g.router.tire.Delete(dp)
	}
}

func (g *group) Handler(fn groupFunction) {
	fn(&RouteHandler{group: g})
}

type RouteHandler struct {
	group *group
}

func (rh *RouteHandler) Route(path ...string) *route {
	return &route{path: path, group: rh.group}
}

func (rh *RouteHandler) Remove(path ...string) {
	if rh.group.router.tire == nil {
		return
	}
	for i := 0; i < len(path); i++ {
		var dp = rh.group.path + path[0]
		if !rh.group.router.StrictMode {
			dp = strings.ToLower(dp)
		}
		rh.group.router.tire.Delete(dp)
	}
}

type route struct {
	path        []string
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

	if len(r.path) == 0 {
		panic("route path can not empty")
	}

	ci := caller.Deep(2)

	var router = r.group.router
	var g = r.group

	if g == nil {
		g = new(group)
	}

	for i := 0; i < len(r.path); i++ {
		var path = router.formatPath(g.path + r.path[0])

		if router.tire == nil {
			router.tire = tire.New[*node]()
		}

		var wba = &node{}

		wba.Info = ci.File + ":" + strconv.Itoa(ci.Line)

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
}

type Router struct {
	tire         *tire.Tire[*node]
	StrictMode   bool
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
	if !r.StrictMode {
		dp = strings.ToLower(dp)
	}
	r.tire.Delete(dp)
}

func (r *Router) Route(path ...string) *route {
	return (&RouteHandler{group: r.Group("")}).Route(path...)
}

func (r *Router) GetAllRouters() []*node {
	var res []*node
	var tires = r.tire.GetAllValue()
	for i := 0; i < len(tires); i++ {
		res = append(res, tires[i].Data)
	}
	return res
}

func (r *Router) getRoute(path string) (*tire.Tire[*node], []byte) {

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
	if !r.StrictMode {
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
