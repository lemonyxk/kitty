/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2022-05-22 03:53
**/

package router

import "strings"

type Group[T any] struct {
	path   string
	desc   []string
	before []Before[T]
	after  []After[T]
	router *Router[T]
}

func (g *Group[T]) Before(before ...Before[T]) *Group[T] {
	g.before = append(g.before, before...)
	return g
}

func (g *Group[T]) After(after ...After[T]) *Group[T] {
	g.after = append(g.after, after...)
	return g
}

func (g *Group[T]) Remove(path ...string) {
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

func (g *Group[T]) Handler(fn func(handler *Handler[T])) {
	fn(&Handler[T]{group: &Group[T]{
		path:   g.path,
		desc:   g.desc,
		before: g.before,
		after:  g.after,
		router: g.router,
	}})
}

func (g *Group[T]) Create() *Handler[T] {
	return &Handler[T]{group: &Group[T]{
		path:   g.path,
		desc:   g.desc,
		before: g.before,
		after:  g.after,
		router: g.router,
	}}
}

func (g *Group[T]) Desc(desc ...string) *Group[T] {
	g.desc = append(g.desc, desc...)
	return g
}
