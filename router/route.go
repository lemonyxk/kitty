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

	"github.com/lemonyxk/caller"
	"github.com/lemonyxk/structure/v3/tire"
)

type Route[T any] struct {
	path        []string
	method      string
	before      []Before[T]
	after       []After[T]
	passBefore  bool
	forceBefore bool
	passAfter   bool
	forceAfter  bool
	group       *Group[T]
}

func (r *Route[T]) Before(before ...Before[T]) *Route[T] {
	r.before = append(r.before, before...)
	return r
}

func (r *Route[T]) PassBefore() *Route[T] {
	r.passBefore = true
	return r
}

func (r *Route[T]) ForceBefore() *Route[T] {
	r.forceBefore = true
	return r
}

func (r *Route[T]) After(after ...After[T]) *Route[T] {
	r.after = append(r.after, after...)
	return r
}

func (r *Route[T]) PassAfter() *Route[T] {
	r.passAfter = true
	return r
}

func (r *Route[T]) ForceAfter() *Route[T] {
	r.forceAfter = true
	return r
}

func (r *Route[T]) Handler(fn Func[T]) {

	if len(r.path) == 0 {
		panic("route path can not empty")
	}

	ci := caller.Deep(2)

	var router = r.group.Router

	var g = r.group

	var method = strings.ToUpper(r.method)

	if g == nil {
		g = new(Group[T])
	}

	for i := 0; i < len(r.path); i++ {
		var path = router.formatPath(g.Path + r.path[i])

		if router.tire == nil {
			router.tire = tire.New[*Node[T]]()
		}

		var cba = &Node[T]{}

		cba.Info = ci.File + ":" + strconv.Itoa(ci.Line)

		cba.Function = fn

		cba.Before = append(g.BeforeList, r.before...)
		if r.passBefore {
			cba.Before = nil
		}
		if r.forceBefore {
			cba.Before = r.before
		}

		cba.After = append(g.AftersList, r.after...)
		if r.passAfter {
			cba.After = nil
		}
		if r.forceAfter {
			cba.After = r.after
		}

		cba.Before = append(cba.Before, router.globalBefore...)
		cba.After = append(cba.After, router.globalAfter...)

		cba.Method = method

		cba.Route = []byte(path)

		router.tire.Insert(path, cba)
	}

}
