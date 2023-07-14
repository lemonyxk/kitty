/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2022-05-22 04:08
**/

package router

import (
	"strconv"
	"strings"
	"unsafe"

	"github.com/lemonyxk/caller"
	"github.com/lemonyxk/structure/trie"
)

type Route[T any] struct {
	path   []string
	method []string
	before []Before[T]
	after  []After[T]
	group  *Group[T]
	desc   []string
}

func (r *Route[T]) Desc(desc ...string) *Route[T] {
	r.desc = append(r.desc, desc...)
	return r
}

func (r *Route[T]) Before(before ...Before[T]) *Route[T] {
	r.before = append(r.before, before...)
	return r
}

func (r *Route[T]) RemoveBefore(before ...Before[T]) *Route[T] {
	for i := 0; i < len(before); i++ {
		for j := 0; j < len(r.before); j++ {
			if *(*unsafe.Pointer)(unsafe.Pointer(&r.before[j])) ==
				*(*unsafe.Pointer)(unsafe.Pointer(&before[i])) {
				r.before = append(r.before[:j], r.before[j+1:]...)
				j--
			}
		}
	}
	return r
}

func (r *Route[T]) CancelBefore() *Route[T] {
	r.before = nil
	return r
}

func (r *Route[T]) After(after ...After[T]) *Route[T] {
	r.after = append(r.after, after...)
	return r
}

func (r *Route[T]) RemoveAfter(after ...After[T]) *Route[T] {
	for i := 0; i < len(after); i++ {
		for j := 0; j < len(r.after); j++ {
			if *(*unsafe.Pointer)(unsafe.Pointer(&r.after[j])) ==
				*(*unsafe.Pointer)(unsafe.Pointer(&after[i])) {
				r.after = append(r.after[:j], r.after[j+1:]...)
				j--
			}
		}
	}
	return r
}

func (r *Route[T]) CancelAfter() *Route[T] {
	r.after = nil
	return r
}

func (r *Route[T]) Handler(fn Func[T]) {

	if len(r.path) == 0 {
		panic("route path can not empty")
	}

	ci := caller.Deep(2)

	var router = r.group.router

	var g = r.group

	var method []string
	for i := 0; i < len(r.method); i++ {
		method = append(method, strings.ToUpper(r.method[i]))
	}

	if g == nil {
		g = new(Group[T])
	}

	for i := 0; i < len(r.path); i++ {

		var originPath = g.path + r.path[i]

		var path = router.formatPath(originPath)

		if router.trie == nil {
			router.trie = trie.New[*Node[T]]()
		}

		var cba = &Node[T]{}

		cba.Info = ci.File + ":" + strconv.Itoa(ci.Line)

		cba.Desc = append(g.desc, r.desc...)

		cba.Function = fn

		cba.Before = append(router.globalBefore, r.before...)
		cba.After = append(router.globalAfter, r.after...)

		cba.Method = method

		cba.Route = []byte(originPath)

		router.trie.Insert(path, cba)
	}

}
