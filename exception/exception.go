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
	"errors"
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

type ErrorFunc func() *Error

func (err ErrorFunc) Error() error {
	if err == nil {
		return nil
	}
	return errors.New(err().Message)
}

type CatchFunc func(ErrorFunc) ErrorFunc

type FinallyFunc func(ErrorFunc) ErrorFunc

var Empty = func() ErrorFunc {
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
	Catch func(CatchFunc) *finally
}

type finally struct {
	Finally   func(FinallyFunc) ErrorFunc
	ErrorFunc func() ErrorFunc
}

func Try(fn func()) (c *catch) {

	defer func() {
		if err := recover(); err != nil {
			var d = 1
			var e = fmt.Errorf("%v", err)
			if strings.HasPrefix(e.Error(), "#assert#") {
				d = 2
			}
			var stacks = NewStackWithError(d, strings.Replace(e.Error(), "#assert#", "", 1))
			c = &catch{Catch: func(f CatchFunc) *finally {
				var ef = f(stacks)
				return &finally{
					Finally: func(ff FinallyFunc) ErrorFunc {
						return ff(ef)
					},
					ErrorFunc: func() ErrorFunc {
						return ef
					},
				}
			}}
		}
	}()

	fn()

	return &catch{Catch: func(f CatchFunc) *finally {
		return &finally{
			Finally: func(ff FinallyFunc) ErrorFunc {
				return ff(nil)
			},
			ErrorFunc: func() ErrorFunc {
				return nil
			},
		}
	}}
}

func Assert(v ...interface{}) {
	if len(v) == 0 {
		return
	}
	if v[len(v)-1] == nil {
		return
	}
	panic(fmt.Errorf("#assert#%v", v[len(v)-1]))
}

func Inspect(v ...interface{}) ErrorFunc {
	if len(v) == 0 {
		return nil
	}
	if v[len(v)-1] == nil {
		return nil
	}
	return newErrorFromDeep(v[len(v)-1], 2)
}

func New(v ...interface{}) ErrorFunc {
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

func newErrorFromDeep(v interface{}, deep int) ErrorFunc {
	if v == nil {
		return nil
	}
	file, line := caller.Caller(deep)
	return newErrorWithFileAndLine(v, file, line)
}

func newErrorWithFileAndLine(v interface{}, file string, line int) ErrorFunc {
	switch v.(type) {
	case error:
		var e ErrorFunc = func() *Error {
			return &Error{time.Now(), file, line, v.(error).Error()}
		}
		return e
	case string:
		var e ErrorFunc = func() *Error {
			return &Error{time.Now(), file, line, v.(string)}
		}
		return e
	default:
		var e ErrorFunc = func() *Error {
			return &Error{time.Now(), file, line, fmt.Sprintf("%v", v)}
		}
		return e
	}
}

func NewStackWithError(deep int, v interface{}) ErrorFunc {
	deep = 10 + deep*2
	var file, line = caller.Stack(deep)
	return newErrorWithFileAndLine(v, file, line)
}
