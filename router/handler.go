/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2022-05-22 04:02
**/

package router

import (
	http2 "net/http"
	"strings"
)

type MethodsHandler[T any,P any] struct {
	method []string
	group  *Group[T,P]
}

func (m *MethodsHandler[T,P]) Route(path ...string) *Route[T,P] {
	return &Route[T,P]{
		method: m.method, path: path, group: m.group,
		before: append([]Before[T]{}, m.group.before...),
		after:  append([]After[T]{}, m.group.after...),
	}
}

type Handler[T any, P any] struct {
	group *Group[T, P]
}

func (rh *Handler[T,P]) Group(path ...string) *Group[T,P] {
	return &Group[T,P]{
		path:   rh.group.path + strings.Join(path, ""),
		desc:   append([]string{}, rh.group.desc...),
		before: append([]Before[T]{}, rh.group.before...),
		after:  append([]After[T]{}, rh.group.after...),
		router: rh.group.router,
	}
}

func (rh *Handler[T,P]) Route(path ...string) *Route[T,P] {
	return &Route[T,P]{
		method: []string{"GET"}, path: path, group: rh.group,
		before: append([]Before[T]{}, rh.group.before...),
		after:  append([]After[T]{}, rh.group.after...),
	}
}

func (rh *Handler[T,P]) Method(method ...string) *MethodsHandler[T,P] {
	return &MethodsHandler[T,P]{method: method, group: rh.group}
}

func (rh *Handler[T,P]) Get(path ...string) *Route[T,P] {
	return rh.Method(http2.MethodGet).Route(path...)
}

func (rh *Handler[T,P]) Post(path ...string) *Route[T,P] {
	return rh.Method(http2.MethodPost).Route(path...)
}

func (rh *Handler[T,P]) Delete(path ...string) *Route[T,P] {
	return rh.Method(http2.MethodDelete).Route(path...)
}

func (rh *Handler[T,P]) Put(path ...string) *Route[T,P] {
	return rh.Method(http2.MethodPut).Route(path...)
}

func (rh *Handler[T,P]) Patch(path ...string) *Route[T,P] {
	return rh.Method(http2.MethodPatch).Route(path...)
}

func (rh *Handler[T,P]) Head(path ...string) *Route[T,P] {
	return rh.Method(http2.MethodHead).Route(path...)
}

func (rh *Handler[T,P]) Options(path ...string) *Route[T,P] {
	return rh.Method(http2.MethodOptions).Route(path...)
}

func (rh *Handler[T,P]) Connect(path ...string) *Route[T,P] {
	return rh.Method(http2.MethodConnect).Route(path...)
}

func (rh *Handler[T,P]) Trace(path ...string) *Route[T,P] {
	return rh.Method(http2.MethodTrace).Route(path...)
}

func (rh *Handler[T,P]) Remove(path ...string) {
	if rh.group.router.trie == nil {
		return
	}
	for i := 0; i < len(path); i++ {
		var dp = rh.group.path + path[i]
		if !rh.group.router.StrictMode {
			dp = strings.ToLower(dp)
		}
		rh.group.router.trie.Delete(dp)
	}
}
