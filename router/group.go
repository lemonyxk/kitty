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

import (
	"strings"
	"unsafe"
)

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

func (g *Group[T]) RemoveBefore(before ...Before[T]) *Group[T] {
	for i := 0; i < len(before); i++ {
		for j := 0; j < len(g.before); j++ {
			if *(*unsafe.Pointer)(unsafe.Pointer(&g.before[j])) ==
				*(*unsafe.Pointer)(unsafe.Pointer(&before[i])) {
				g.before = append(g.before[:j], g.before[j+1:]...)
				j--
			}
		}
	}
	return g
}

func (g *Group[T]) CancelBefore() *Group[T] {
	g.before = nil
	return g
}

func (g *Group[T]) RemoveAfter(after ...After[T]) *Group[T] {
	for i := 0; i < len(after); i++ {
		for j := 0; j < len(g.after); j++ {
			if *(*unsafe.Pointer)(unsafe.Pointer(&g.after[j])) ==
				*(*unsafe.Pointer)(unsafe.Pointer(&after[i])) {
				g.after = append(g.after[:j], g.after[j+1:]...)
				j--
			}
		}
	}
	return g
}

func (g *Group[T]) CancelAfter() *Group[T] {
	g.after = nil
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
		desc:   append([]string{}, g.desc...),
		before: append([]Before[T]{}, g.before...),
		after:  append([]After[T]{}, g.after...),
		router: g.router,
	}})
}

func (g *Group[T]) Create() *Handler[T] {
	return &Handler[T]{group: &Group[T]{
		path:   g.path,
		desc:   append([]string{}, g.desc...),
		before: append([]Before[T]{}, g.before...),
		after:  append([]After[T]{}, g.after...),
		router: g.router,
	}}
}

func (g *Group[T]) Desc(desc ...string) *Group[T] {
	g.desc = append(g.desc, desc...)
	return g
}
