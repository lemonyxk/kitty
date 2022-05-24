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
	"net/http"
)

// Router

type Static struct {
	fileSystem http.FileSystem
	prefixPath string
	fixPath    string
	index      int
}

type StaticRouter struct {
	static                 []*Static
	defaultIndex           []string
	staticFileMiddle       map[string]func(w http.ResponseWriter, r *http.Request, f http.File, i fs.FileInfo) error
	staticGlobalFileMiddle func(w http.ResponseWriter, r *http.Request, f http.File, i fs.FileInfo) error
	staticDirMiddle        map[string]func(w http.ResponseWriter, r *http.Request, f http.File, i fs.FileInfo) error
	staticGlobalDirMiddle  func(w http.ResponseWriter, r *http.Request, f http.File, i fs.FileInfo) error
	staticDownload         bool
	openDir                []int
}

func (r *StaticRouter) SetDefaultIndex(index ...string) {
	r.defaultIndex = index
}

func (r *StaticRouter) SetOpenDir(dirIndex ...int) {
	r.openDir = dirIndex
}

func (r *StaticRouter) SetStaticDownload(flag bool) {
	r.staticDownload = flag
}

func (r *StaticRouter) SetStaticFileMiddle(t ...string) *StaticFileMiddle {
	return &StaticFileMiddle{r, t}
}

func (r *StaticRouter) SetStaticDirMiddle(t ...string) *StaticDirMiddle {
	return &StaticDirMiddle{r, t}
}

func (r *StaticRouter) SetStaticGlobalFileMiddle(fn func(w http.ResponseWriter, r *http.Request, f http.File, i fs.FileInfo) error) {
	r.staticGlobalFileMiddle = fn
}

func (r *StaticRouter) SetStaticGlobalDirMiddle(fn func(w http.ResponseWriter, r *http.Request, f http.File, i fs.FileInfo) error) {
	r.staticGlobalDirMiddle = fn
}

func (r *StaticRouter) SetStaticPath(prefixPath string, fixPath string, fileSystem http.FileSystem) int {

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
	r.staticFileMiddle = make(map[string]func(w http.ResponseWriter, r *http.Request, f http.File, i fs.FileInfo) error)
	r.staticDirMiddle = make(map[string]func(w http.ResponseWriter, r *http.Request, f http.File, i fs.FileInfo) error)

	return static.index
}

// StaticFileMiddle

type StaticFileMiddle struct {
	r *StaticRouter
	t []string
}

func (s *StaticFileMiddle) Handler(fn func(w http.ResponseWriter, r *http.Request, f http.File, i fs.FileInfo) error) {
	for i := 0; i < len(s.t); i++ {
		s.r.staticFileMiddle[s.t[i]] = fn
	}
}

// StaticDirMiddle

type StaticDirMiddle struct {
	r *StaticRouter
	t []string
}

func (s *StaticDirMiddle) Handler(fn func(w http.ResponseWriter, r *http.Request, f http.File, i fs.FileInfo) error) {
	for i := 0; i < len(s.t); i++ {
		s.r.staticDirMiddle[s.t[i]] = fn
	}
}
