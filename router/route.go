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

type Route[T any, P any] struct {
	path   []string
	method []string
	before []Before[T]
	after  []After[T]
	group  *Group[T, P]
	desc   []string
	data   P
}

func (r *Route[T, P]) Desc(desc ...string) *Route[T, P] {
	r.desc = append(r.desc, desc...)
	return r
}

func (r *Route[T, P]) Before(before ...Before[T]) *Route[T, P] {
	r.before = append(r.before, before...)
	return r
}

func (r *Route[T, P]) RemoveBefore(before ...Before[T]) *Route[T, P] {
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

func (r *Route[T, P]) CancelBefore() *Route[T, P] {
	r.before = nil
	return r
}

func (r *Route[T, P]) After(after ...After[T]) *Route[T, P] {
	r.after = append(r.after, after...)
	return r
}

func (r *Route[T, P]) RemoveAfter(after ...After[T]) *Route[T, P] {
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

func (r *Route[T, P]) CancelAfter() *Route[T, P] {
	r.after = nil
	return r
}

func (r *Route[T, P]) Data(data P) *Route[T, P] {
	r.data = data
	return r
}

func (r *Route[T, P]) Handler(fn Func[T]) {

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
		g = new(Group[T, P])
	}

	for i := 0; i < len(r.path); i++ {

		var originPath = g.path + r.path[i]

		var path = router.formatPath(originPath)

		if router.trie == nil {
			router.trie = trie.New[*Node[T, P]]()
		}

		var cba = &Node[T, P]{}

		cba.Info = ci.File + ":" + strconv.Itoa(ci.Line)

		cba.Desc = append(g.desc, r.desc...)

		cba.Function = fn

		cba.Before = append(router.globalBefore, r.before...)
		cba.After = append(router.globalAfter, r.after...)

		cba.Method = method

		cba.Route = []byte(originPath)

		cba.Data = r.data

		router.trie.Insert(path, cba)
	}

}
