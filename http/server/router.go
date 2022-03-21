/**
* @program: lemon
*
* @description:
*
* @author: lemon
*
* @create: 2019-11-25 11:29
**/

package server

import (
	"io/fs"
	http2 "net/http"
	"strconv"
	"strings"

	"github.com/lemonyxk/caller"
	"github.com/lemonyxk/structure/v3/tire"

	"github.com/lemonyxk/kitty/v2/http"
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

func (g *group) Remove(path ...string) {
	if g.router.tire == nil {
		return
	}
	for i := 0; i < len(path); i++ {
		var dp = g.path + path[i]
		if !g.router.StrictMode {
			dp = strings.ToLower(dp)
		}
		g.router.tire.Delete(dp)
	}
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

func (rh *RouteHandler) Remove(path ...string) {
	if rh.group.router.tire == nil {
		return
	}
	for i := 0; i < len(path); i++ {
		var dp = rh.group.path + path[i]
		if !rh.group.router.StrictMode {
			dp = strings.ToLower(dp)
		}
		rh.group.router.tire.Delete(dp)
	}
}

func (rh *RouteHandler) Route(method string, path ...string) *route {
	return &route{path: path, method: method, group: rh.group}
}

func (rh *RouteHandler) Get(path ...string) *route {
	return rh.Route("GET", path...)
}

func (rh *RouteHandler) Post(path ...string) *route {
	return rh.Route("POST", path...)
}

func (rh *RouteHandler) Delete(path ...string) *route {
	return rh.Route("DELETE", path...)
}

func (rh *RouteHandler) Put(path ...string) *route {
	return rh.Route("PUT", path...)
}

func (rh *RouteHandler) Patch(path ...string) *route {
	return rh.Route("PATCH", path...)
}

func (rh *RouteHandler) Option(path ...string) *route {
	return rh.Route("OPTION", path...)
}

type route struct {
	path        []string
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

	if len(r.path) == 0 || r.method == "" {
		panic("route path or method can not empty")
	}

	ci := caller.Deep(2)

	var g = r.group

	var router = r.group.router

	var method = strings.ToUpper(r.method)

	if g == nil {
		g = new(group)
	}

	for i := 0; i < len(r.path); i++ {

		var path = router.formatPath(g.path + r.path[i])

		if router.tire == nil {
			router.tire = tire.New[*node]()
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
}

type Static struct {
	fileSystem http2.FileSystem
	prefixPath string
	fixPath    string
	index      int
}

type Router struct {
	StrictMode   bool
	tire         *tire.Tire[*node]
	globalAfter  []After
	globalBefore []Before

	static                 []*Static
	defaultIndex           []string
	staticFileMiddle       map[string]func(w http2.ResponseWriter, r *http2.Request, f http2.File, i fs.FileInfo) error
	staticGlobalFileMiddle func(w http2.ResponseWriter, r *http2.Request, f http2.File, i fs.FileInfo) error
	staticDirMiddle        map[string]func(w http2.ResponseWriter, r *http2.Request, f http2.File, i fs.FileInfo) error
	staticGlobalDirMiddle  func(w http2.ResponseWriter, r *http2.Request, f http2.File, i fs.FileInfo) error
	staticDownload         bool
	openDir                []int
}

func (r *Router) SetGlobalBefore(before ...Before) {
	r.globalBefore = append(r.globalBefore, before...)
}

func (r *Router) SetGlobalAfter(after ...After) {
	r.globalAfter = append(r.globalAfter, after...)
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

func (r *Router) GetAllRouters() []*node {
	var res []*node
	var tires = r.tire.GetAllValue()
	for i := 0; i < len(tires); i++ {
		res = append(res, tires[i].Data)
	}
	return res
}

func (r *Router) SetDefaultIndex(index ...string) {
	r.defaultIndex = index
}

func (r *Router) SetOpenDir(dirIndex ...int) {
	r.openDir = dirIndex
}

func (r *Router) SetStaticDownload(flag bool) {
	r.staticDownload = flag
}

type StaticFileMiddle struct {
	r *Router
	t []string
}

func (s *StaticFileMiddle) Handler(fn func(w http2.ResponseWriter, r *http2.Request, f http2.File, i fs.FileInfo) error) {
	for i := 0; i < len(s.t); i++ {
		s.r.staticFileMiddle[s.t[i]] = fn
	}
}

func (r *Router) SetStaticFileMiddle(t ...string) *StaticFileMiddle {
	return &StaticFileMiddle{r, t}
}

type StaticDirMiddle struct {
	r *Router
	t []string
}

func (s *StaticDirMiddle) Handler(fn func(w http2.ResponseWriter, r *http2.Request, f http2.File, i fs.FileInfo) error) {
	for i := 0; i < len(s.t); i++ {
		s.r.staticDirMiddle[s.t[i]] = fn
	}
}

func (r *Router) SetStaticDirMiddle(t ...string) *StaticDirMiddle {
	return &StaticDirMiddle{r, t}
}

func (r *Router) SetStaticGlobalFileMiddle(fn func(w http2.ResponseWriter, r *http2.Request, f http2.File, i fs.FileInfo) error) {
	r.staticGlobalFileMiddle = fn
}

func (r *Router) SetStaticGlobalDirMiddle(fn func(w http2.ResponseWriter, r *http2.Request, f http2.File, i fs.FileInfo) error) {
	r.staticGlobalDirMiddle = fn
}

func (r *Router) SetStaticPath(prefixPath string, fixPath string, fileSystem http2.FileSystem) int {

	if prefixPath == "" {
		panic("prefixPath can not be empty")
	}

	if fileSystem == nil {
		panic("fileSystem can not be empty")
	}

	for i := 0; i < len(r.static); i++ {
		if r.static[i].prefixPath == prefixPath {
			panic("prefixPath is exist")
		}
	}

	var static = &Static{fileSystem, prefixPath, fixPath, len(r.static)}
	r.static = append(r.static, static)
	r.staticFileMiddle = make(map[string]func(w http2.ResponseWriter, r *http2.Request, f http2.File, i fs.FileInfo) error)
	r.staticDirMiddle = make(map[string]func(w http2.ResponseWriter, r *http2.Request, f http2.File, i fs.FileInfo) error)

	return static.index
}

func (r *Router) Group(path ...string) *group {

	var g = new(group)

	g.path = strings.Join(path, "")

	g.router = r

	return g
}

func (r *Router) Route(method string, path ...string) *route {
	return (&RouteHandler{group: r.Group("")}).Route(method, path...)
}

func (r *Router) getRoute(method string, path string) (*tire.Tire[*node], []byte) {

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

	if t.Data.Method != method {
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
