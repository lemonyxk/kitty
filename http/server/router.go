/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-11-25 11:29
**/

package server

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/lemoyxk/caller"
	"github.com/lemoyxk/structure/tire"

	"github.com/lemoyxk/kitty/http"
)

type groupFunction func(handler *RouteHandler)

type function func(stream *http.Stream) error

type Before func(stream *http.Stream) error

type After func(stream *http.Stream) error

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

func (g *group) Remove(path string) {
	g.router.tire.Delete(g.path + path)
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

func (rh *RouteHandler) Remove(path string) {
	rh.group.router.tire.Delete(rh.group.path + path)
}

func (rh *RouteHandler) Route(method string, path string) *route {
	return &route{path: path, method: method, group: rh.group}
}

func (rh *RouteHandler) Get(path string) *route {
	return rh.Route("GET", path)
}

func (rh *RouteHandler) Post(path string) *route {
	return rh.Route("POST", path)
}

func (rh *RouteHandler) Delete(path string) *route {
	return rh.Route("DELETE", path)
}

func (rh *RouteHandler) Put(path string) *route {
	return rh.Route("PUT", path)
}

func (rh *RouteHandler) Patch(path string) *route {
	return rh.Route("PATCH", path)
}

func (rh *RouteHandler) Option(path string) *route {
	return rh.Route("OPTION", path)
}

type route struct {
	path        string
	method      string
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

	if r.path == "" || r.method == "" {
		panic("route path or method can not empty")
	}

	ci := caller.Deep(2)

	var g = r.group

	var router = r.group.router

	var method = strings.ToUpper(r.method)

	if g == nil {
		g = new(group)
	}

	var path = router.formatPath(g.path + r.path)

	if router.tire == nil {
		router.tire = new(tire.Tire)
	}

	var hba = &node{}

	hba.Info = ci.File + ":" + strconv.Itoa(ci.Line)

	hba.Function = fn

	hba.Before = append(g.before, r.before...)
	if r.passBefore {
		hba.Before = nil
	}
	if r.forceBefore {
		hba.Before = r.before
	}

	hba.After = append(g.after, r.after...)
	if r.passAfter {
		hba.After = nil
	}
	if r.forceAfter {
		hba.After = r.after
	}

	hba.Before = append(hba.Before, router.globalBefore...)
	hba.After = append(hba.After, router.globalAfter...)

	hba.Method = method

	hba.Route = []byte(path)

	router.tire.Insert(path, hba)
}

type Router struct {
	StrictMode   bool
	tire         *tire.Tire
	prefixPath   string
	staticPath   string
	defaultIndex string
	globalAfter  []After
	globalBefore []Before
}

func (r *Router) SetGlobalBefore(before ...Before) {
	r.globalBefore = append(r.globalBefore, before...)
}

func (r *Router) SetGlobalAfter(after ...After) {
	r.globalAfter = append(r.globalAfter, after...)
}

func (r *Router) Remove(path ...string) {
	r.tire.Delete(strings.Join(path, ""))
}

func (r *Router) GetAllRouters() []*node {
	var res []*node
	var tires = r.tire.GetAllValue()
	for i := 0; i < len(tires); i++ {
		res = append(res, tires[i].Data.(*node))
	}
	return res
}

func (r *Router) SetDefaultIndex(index string) {
	r.defaultIndex = index
}

func (r *Router) SetStaticPath(prefixPath string, staticPath string) {

	if prefixPath == "" {
		panic("prefixPath can not be empty")
	}

	if staticPath == "" {
		panic("staticPath can not be empty")
	}

	absStaticPath, err := filepath.Abs(staticPath)
	if err != nil {
		panic(err)
	}

	info, err := os.Stat(absStaticPath)
	if err != nil {
		panic(err)
	}

	if !info.IsDir() {
		panic("staticPath is not a dir")
	}

	r.prefixPath = prefixPath
	r.staticPath = absStaticPath
	r.defaultIndex = "index.html"
}

func (r *Router) Group(path ...string) *group {

	var g = new(group)

	g.path = strings.Join(path, "")

	g.router = r

	return g
}

func (r *Router) Route(method string, path string) *route {
	return (&RouteHandler{group: r.Group("")}).Route(method, path)
}

func (r *Router) getRoute(method string, path string) (*tire.Tire, []byte) {

	if r.tire == nil {
		return nil, nil
	}

	method = strings.ToUpper(method)

	path = r.formatPath(path)

	var pathB = []byte(path)

	var t = r.tire.GetValue(pathB)

	if t == nil {
		return nil, nil
	}

	if t.Data.(*node).Method != method {
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
	Method   string
	Function function
	Before   []Before
	After    []After
}
