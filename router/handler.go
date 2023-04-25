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

// type GroupFunc[T any] func(handler *RouteHandler[T])

type MethodsHandler[T any] struct {
	method []string
	group  *Group[T]
}

func (m *MethodsHandler[T]) Route(path ...string) *Route[T] {
	return &Route[T]{
		method: m.method, path: path, group: m.group,
		before: append([]Before[T]{}, m.group.before...),
		after:  append([]After[T]{}, m.group.after...),
	}
}

type Handler[T any] struct {
	group *Group[T]
}

func (rh *Handler[T]) Group(path ...string) *Group[T] {
	return &Group[T]{
		path:   rh.group.path + strings.Join(path, ""),
		desc:   append([]string{}, rh.group.desc...),
		before: append([]Before[T]{}, rh.group.before...),
		after:  append([]After[T]{}, rh.group.after...),
		router: rh.group.router,
	}
}

func (rh *Handler[T]) Route(path ...string) *Route[T] {
	return &Route[T]{
		method: []string{"GET"}, path: path, group: rh.group,
		before: append([]Before[T]{}, rh.group.before...),
		after:  append([]After[T]{}, rh.group.after...),
	}
}

func (rh *Handler[T]) Method(method ...string) *MethodsHandler[T] {
	return &MethodsHandler[T]{method: method, group: rh.group}
}

func (rh *Handler[T]) Get(path ...string) *Route[T] {
	return rh.Method(http2.MethodGet).Route(path...)
}

func (rh *Handler[T]) Post(path ...string) *Route[T] {
	return rh.Method(http2.MethodPost).Route(path...)
}

func (rh *Handler[T]) Delete(path ...string) *Route[T] {
	return rh.Method(http2.MethodDelete).Route(path...)
}

func (rh *Handler[T]) Put(path ...string) *Route[T] {
	return rh.Method(http2.MethodPut).Route(path...)
}

func (rh *Handler[T]) Patch(path ...string) *Route[T] {
	return rh.Method(http2.MethodPatch).Route(path...)
}

func (rh *Handler[T]) Head(path ...string) *Route[T] {
	return rh.Method(http2.MethodHead).Route(path...)
}

func (rh *Handler[T]) Options(path ...string) *Route[T] {
	return rh.Method(http2.MethodOptions).Route(path...)
}

func (rh *Handler[T]) Connect(path ...string) *Route[T] {
	return rh.Method(http2.MethodConnect).Route(path...)
}

func (rh *Handler[T]) Trace(path ...string) *Route[T] {
	return rh.Method(http2.MethodTrace).Route(path...)
}

func (rh *Handler[T]) Remove(path ...string) {
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
