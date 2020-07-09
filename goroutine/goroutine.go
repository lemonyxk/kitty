/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-12-19 19:17
**/

package goroutine

import (
	"fmt"
	"strings"
	"time"

	"github.com/Lemo-yxk/structure/queue"

	"github.com/Lemo-yxk/lemo/exception"
	"github.com/Lemo-yxk/lemo/utils"
)

var que = queue.NewBlockQueue()

func Run(fn func()) {
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
				var file, line = utils.Stack.Stack(d)

				que.Push(exception.NewException(time.Now(), file, line, s))
			}
		}()
		fn()
	}()
}

func Watch(fn func(err exception.Error)) {
	go func() {
		for {
			fn(que.Pop().(exception.Error))
		}
	}()
}
