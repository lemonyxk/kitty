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
	return &Route[T]{method: m.method, path: path, group: m.group}
}

type Handler[T any] struct {
	group *Group[T]
}

func (rh *Handler[T]) Group(path ...string) *Group[T] {
	return &Group[T]{
		Path:        rh.group.Path + strings.Join(path, ""),
		Description: rh.group.Description,
		BeforeList:  rh.group.BeforeList,
		AftersList:  rh.group.AftersList,
		Router:      rh.group.Router,
	}
}

func (rh *Handler[T]) Route(path ...string) *Route[T] {
	return &Route[T]{method: []string{"GET"}, path: path, group: rh.group}
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
	if rh.group.Router.tire == nil {
		return
	}
	for i := 0; i < len(path); i++ {
		var dp = rh.group.Path + path[i]
		if !rh.group.Router.StrictMode {
			dp = strings.ToLower(dp)
		}
		rh.group.Router.tire.Delete(dp)
	}
}
