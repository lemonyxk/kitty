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

	"github.com/lemonyxk/structure/tire"
)

type Router[T any] struct {
	StrictMode   bool
	tire         *tire.Tire[*Node[T]]
	globalAfter  []After[T]
	globalBefore []Before[T]
}

func (r *Router[T]) SetGlobalBefore(before ...Before[T]) {
	r.globalBefore = append(r.globalBefore, before...)
}

func (r *Router[T]) SetGlobalAfter(after ...After[T]) {
	r.globalAfter = append(r.globalAfter, after...)
}

func (r *Router[T]) GetAllRouters() []*Node[T] {
	var res []*Node[T]
	var tires = r.tire.GetAllValue()
	for i := 0; i < len(tires); i++ {
		res = append(res, tires[i].Data)
	}
	return res
}

func (r *Router[T]) Group(path ...string) *Group[T] {

	var g = new(Group[T])

	g.path = strings.Join(path, "")

	g.router = r

	return g
}

func (r *Router[T]) Remove(path ...string) {
	if r.tire == nil {
		return
	}
	var dp = strings.Join(path, "")
	if !r.StrictMode {
		dp = strings.ToLower(dp)
	}
	r.tire.Delete(dp)
}

func (r *Router[T]) Create() *Handler[T] {
	return &Handler[T]{group: r.Group("")}
}

func (r *Router[T]) Route(path ...string) *Route[T] {
	return (&Handler[T]{group: r.Group("")}).Route(path...)
}

func (r *Router[T]) Method(method ...string) *MethodsRouter[T] {
	return &MethodsRouter[T]{router: r, method: method}
}

type MethodsRouter[T any] struct {
	router *Router[T]
	method []string
}

func (m *MethodsRouter[T]) Route(path ...string) *Route[T] {
	return (&Handler[T]{group: m.router.Group("")}).Method(m.method...).Route(path...)
}

func (r *Router[T]) GetRoute(path string) (*tire.Tire[*Node[T]], []byte) {

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

func (r *Router[T]) formatPath(path string) string {
	if !r.StrictMode {
		path = strings.ToLower(path)
	}
	return path
}
