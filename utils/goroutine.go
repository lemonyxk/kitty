/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-12-19 19:17
**/

package utils

import (
	"fmt"
	"strings"

	"github.com/Lemo-yxk/lemo/container/queue"
	"github.com/Lemo-yxk/lemo/exception"
)

type goroutine int

const Goroutine goroutine = iota

var que = queue.NewBlockQueue()

func (g goroutine) Run(fn func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				var d = 1
				var e = fmt.Errorf("%v", err)
				if strings.HasPrefix(e.Error(), "#assert#") {
					d = 2
				}
				que.Push(exception.NewStackWithError(d, strings.Replace(e.Error(), "#assert#", "", 1)))
			}
		}()
		fn()
	}()
}

func (g goroutine) Watch(fn func(exception.ErrorFunc)) {
	go func() {
		for {
			fn(que.Pop().(exception.ErrorFunc))
		}
	}()
}
