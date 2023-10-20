/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2022-05-22 03:58
**/

package router

import (
	"strings"

	"github.com/lemonyxk/structure/trie"
)

type Router[T any, P any] struct {
	StrictMode   bool
	trie         *trie.Node[*Node[T, P]]
	globalAfter  []After[T]
	globalBefore []Before[T]
}

func (r *Router[T, P]) SetGlobalBefore(before ...Before[T]) {
	r.globalBefore = append(r.globalBefore, before...)
}

func (r *Router[T, P]) SetGlobalAfter(after ...After[T]) {
	r.globalAfter = append(r.globalAfter, after...)
}

func (r *Router[T, P]) GetAllRouters() []*Node[T, P] {
	var res []*Node[T, P]
	var tires = r.trie.GetAllValue()
	for i := 0; i < len(tires); i++ {
		res = append(res, tires[i].Data)
	}
	return res
}

func (r *Router[T, P]) Group(path ...string) *Group[T, P] {
	var g = new(Group[T, P])
	g.path = strings.Join(path, "")
	g.router = r
	return g
}

func (r *Router[T, P]) Remove(path ...string) {
	if r.trie == nil {
		return
	}
	var dp = strings.Join(path, "")
	if !r.StrictMode {
		dp = strings.ToLower(dp)
	}
	r.trie.Delete(dp)
}

func (r *Router[T, P]) Create() *Handler[T, P] {
	return &Handler[T, P]{group: r.Group("")}
}

func (r *Router[T, P]) Route(path ...string) *Route[T, P] {
	return (&Handler[T, P]{group: r.Group("")}).Route(path...)
}

func (r *Router[T, P]) Method(method ...string) *MethodsRouter[T, P] {
	return &MethodsRouter[T, P]{router: r, method: method}
}

type MethodsRouter[T any, P any] struct {
	router *Router[T, P]
	method []string
}

func (m *MethodsRouter[T, P]) Route(path ...string) *Route[T, P] {
	return (&Handler[T, P]{group: m.router.Group("")}).Method(m.method...).Route(path...)
}

func (r *Router[T, P]) GetRoute(path string) (*trie.Node[*Node[T, P]], string) {
	if r.trie == nil {
		return nil, ""
	}

	path = r.formatPath(path)

	var t = r.trie.GetValue(path)
	if t == nil {
		return nil, ""
	}

	return t, path
}

func (r *Router[T, P]) formatPath(path string) string {
	if !r.StrictMode {
		path = strings.ToLower(path)
	}
	return path
}
