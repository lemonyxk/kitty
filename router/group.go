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

type Group[T any, P any] struct {
	path   string
	desc   []string
	data   P
	before []Before[T]
	after  []After[T]
	router *Router[T, P]
}

func (g *Group[T, P]) Before(before ...Before[T]) *Group[T, P] {
	g.before = append(g.before, before...)
	return g
}

func (g *Group[T, P]) After(after ...After[T]) *Group[T, P] {
	g.after = append(g.after, after...)
	return g
}

func (g *Group[T, P]) RemoveBefore(before ...Before[T]) *Group[T, P] {
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

func (g *Group[T, P]) CancelBefore() *Group[T, P] {
	g.before = nil
	return g
}

func (g *Group[T, P]) RemoveAfter(after ...After[T]) *Group[T, P] {
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

func (g *Group[T, P]) CancelAfter() *Group[T, P] {
	g.after = nil
	return g
}

func (g *Group[T, P]) Remove(path ...string) {
	if g.router.trie == nil {
		return
	}
	for i := 0; i < len(path); i++ {
		var dp = g.path + path[i]
		if !g.router.StrictMode {
			dp = strings.ToLower(dp)
		}
		g.router.trie.Delete(dp)
	}
}

func (g *Group[T, P]) Handler(fn func(handler *Handler[T, P])) {
	fn(&Handler[T, P]{group: &Group[T, P]{
		path:   g.path,
		desc:   append([]string{}, g.desc...),
		before: append([]Before[T]{}, g.before...),
		after:  append([]After[T]{}, g.after...),
		router: g.router,
	}})
}

func (g *Group[T, P]) Create() *Handler[T, P] {
	return &Handler[T, P]{group: &Group[T, P]{
		path:   g.path,
		desc:   append([]string{}, g.desc...),
		before: append([]Before[T]{}, g.before...),
		after:  append([]After[T]{}, g.after...),
		router: g.router,
	}}
}

func (g *Group[T, P]) Desc(desc ...string) *Group[T, P] {
	g.desc = append(g.desc, desc...)
	return g
}
