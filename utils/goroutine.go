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
	"time"

	"github.com/Lemo-yxk/lemo/caller"
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
				var d = 0
				var e = fmt.Errorf("%v", err)
				var s = e.Error()
				if strings.HasPrefix(s, "#exception#") {
					d = 1
					s = strings.Replace(s, "#exception#", "", 1)
				}

				d = 10 + d*2
				var file, line = caller.Stack(d)

				que.Push(exception.NewException(time.Now(), file, line, s))
			}
		}()
		fn()
	}()
}

func (g goroutine) Watch(fn func(err exception.Error)) {
	go func() {
		for {
			fn(que.Pop().(exception.Error))
		}
	}()
}
