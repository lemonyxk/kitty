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

type Handler[T any] struct {
	group *Group[T]
}

func (rh *Handler[T]) Route(path ...string) *Route[T] {
	return &Route[T]{method: "GET", path: path, group: rh.group}
}

func (rh *Handler[T]) RouteMethod(method string, path ...string) *Route[T] {
	return &Route[T]{method: method, path: path, group: rh.group}
}

func (rh *Handler[T]) Get(path ...string) *Route[T] {
	return rh.RouteMethod("GET", path...)
}

func (rh *Handler[T]) Post(path ...string) *Route[T] {
	return rh.RouteMethod("POST", path...)
}

func (rh *Handler[T]) Delete(path ...string) *Route[T] {
	return rh.RouteMethod("DELETE", path...)
}

func (rh *Handler[T]) Put(path ...string) *Route[T] {
	return rh.RouteMethod("PUT", path...)
}

func (rh *Handler[T]) Patch(path ...string) *Route[T] {
	return rh.RouteMethod("PATCH", path...)
}

func (rh *Handler[T]) Option(path ...string) *Route[T] {
	return rh.RouteMethod("OPTION", path...)
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
