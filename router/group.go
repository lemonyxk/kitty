/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2022-05-22 03:53
**/

package router

import "strings"

type Group[T any] struct {
	Path       string
	BeforeList []Before[T]
	AftersList []After[T]
	Router     *Router[T]
}

func (g *Group[T]) Before(before ...Before[T]) *Group[T] {
	g.BeforeList = append(g.BeforeList, before...)
	return g
}

func (g *Group[T]) After(after ...After[T]) *Group[T] {
	g.AftersList = append(g.AftersList, after...)
	return g
}

func (g *Group[T]) Remove(path ...string) {
	if g.Router.tire == nil {
		return
	}
	for i := 0; i < len(path); i++ {
		var dp = g.Path + path[i]
		if !g.Router.StrictMode {
			dp = strings.ToLower(dp)
		}
		g.Router.tire.Delete(dp)
	}
}

func (g *Group[T]) Handler(fn func(handler *Handler[T])) {
	fn(&Handler[T]{group: g})
}

func (g *Group[T]) Create() *Handler[T] {
	return &Handler[T]{group: g}
}
