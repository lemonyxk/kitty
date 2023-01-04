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

import "strings"

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

func (rh *Handler[T]) Route(path ...string) *Route[T] {
	return &Route[T]{method: []string{"GET"}, path: path, group: rh.group}
}

func (rh *Handler[T]) Method(method ...string) *MethodsHandler[T] {
	return &MethodsHandler[T]{method: method, group: rh.group}
}

func (rh *Handler[T]) Get(path ...string) *Route[T] {
	return rh.Method("GET").Route(path...)
}

func (rh *Handler[T]) Post(path ...string) *Route[T] {
	return rh.Method("POST").Route(path...)
}

func (rh *Handler[T]) Delete(path ...string) *Route[T] {
	return rh.Method("DELETE").Route(path...)
}

func (rh *Handler[T]) Put(path ...string) *Route[T] {
	return rh.Method("PUT").Route(path...)
}

func (rh *Handler[T]) Patch(path ...string) *Route[T] {
	return rh.Method("PATCH").Route(path...)
}

func (rh *Handler[T]) Option(path ...string) *Route[T] {
	return rh.Method("OPTION").Route(path...)
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
