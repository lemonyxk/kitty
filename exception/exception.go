/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-09-25 20:37
**/

package exception

import (
	"fmt"
	"strings"
	"time"

	"github.com/Lemo-yxk/lemo/caller"
)

type Error struct {
	Time    time.Time
	File    string
	Line    int
	Message string
}

type CatchFn func(err error, trace *caller.Trace) func() *Error

var Empty = func() func() *Error {
	return func() *Error {
		return nil
	}
}

func (err *Error) Error() string {
	return err.Message
}

func (err *Error) String() string {
	return fmt.Sprintf("ERR %s %s:%d %s", err.Time.Format("2006-01-02 15:04:05"), err.File, err.Line, err.Message)
}

type catch struct {
	Catch func(CatchFn) func() *Error
}

func Try(fn func()) (c *catch) {
	defer func() {
		if err := recover(); err != nil {
			var traces = caller.Stack()
			c = &catch{Catch: func(f CatchFn) func() *Error {
				return f(fmt.Errorf("%v", err), traces)
			}}
		}
	}()

	fn()

	return &catch{Catch: func(f CatchFn) func() *Error {
		return f(nil, nil)
	}}
}

func Assert(v ...interface{}) {
	if len(v) == 0 {
		return
	}
	if v[len(v)-1] == nil {
		return
	}
	panic(v[len(v)-1])
}

func Inspect(v ...interface{}) func() *Error {
	if len(v) == 0 {
		return nil
	}
	if v[len(v)-1] == nil {
		return nil
	}
	return newErrorFromDeep(v[len(v)-1], 2)
}

func New(v ...interface{}) func() *Error {
	if len(v) == 0 {
		return nil
	}

	var invalid = true
	for i := 0; i < len(v); i++ {
		if v[i] != nil {
			invalid = false
			break
		}
	}

	if invalid {
		return nil
	}

	if len(v) == 1 {
		return newErrorFromDeep(v[0], 2)
	}

	if len(v) > 1 {
		if format, ok := v[0].(string); ok && len(format) > 1 && strings.Index(format, "%") != -1 {
			return newErrorFromDeep(fmt.Errorf(format, v[1:]...), 2)
		}
	}

	var str = fmt.Sprintln(v...)
	return newErrorFromDeep(str[:len(str)-1], 2)
}

func newErrorFromDeep(v interface{}, deep int) func() *Error {

	if v == nil {
		return nil
	}

	file, line := caller.Caller(deep)

	switch v.(type) {
	case error:
		return func() *Error {
			return &Error{time.Now(), file, line, v.(error).Error()}
		}
	case string:
		return func() *Error {
			return &Error{time.Now(), file, line, v.(string)}
		}
	default:
		return func() *Error {
			return &Error{time.Now(), file, line, fmt.Sprintf("%v", v)}
		}
	}

}
