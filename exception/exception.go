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
	"reflect"
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
			if strings.HasPrefix(e.Error(), "#exception#") {
				d = 2
			}
			var stacks = NewStackWithError(d, strings.Replace(e.Error(), "#exception#", "", 1))
			c = &catch{Catch: func(f CatchFunc) *finally {
				var ef = f(stacks)
				return &finally{
					Finally:   func(ff FinallyFunc) ErrorFunc { return ff(ef) },
					ErrorFunc: func() ErrorFunc { return ef },
				}
			}}
		}
	}()

	fn()

	return &catch{Catch: func(f CatchFunc) *finally {
		return &finally{
			Finally:   func(ff FinallyFunc) ErrorFunc { return ff(nil) },
			ErrorFunc: func() ErrorFunc { return nil },
		}
	}}
}

func Throw(v interface{}) {
	panic(fmt.Errorf("#exception#%v", v))
}

func Assert(v ...interface{}) {
	if len(v) == 0 {
		return
	}
	if IsNil(v[len(v)-1]) {
		return
	}
	panic(fmt.Errorf("#exception#%v", v[len(v)-1]))
}

func Inspect(v ...interface{}) ErrorFunc {
	if len(v) == 0 {
		return nil
	}
	if IsNil(v[len(v)-1]) {
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
		if !IsNil(v[i]) {
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

	var str = fmt.Sprintln(v...)
	return newErrorFromDeep(str[:len(str)-1], 2)
}

func NewFormat(format string, v ...interface{}) ErrorFunc {
	if len(v) == 0 {
		return nil
	}

	var invalid = true
	for i := 0; i < len(v); i++ {
		if !IsNil(v[i]) {
			invalid = false
			break
		}
	}

	if invalid {
		return nil
	}

	var str = fmt.Sprintf(format, v...)
	return newErrorFromDeep(str, 2)
}

func newErrorFromDeep(v interface{}, deep int) ErrorFunc {
	file, line := caller.Caller(deep)
	return newErrorWithFileAndLine(v, file, line)
}

func newErrorWithFileAndLine(v interface{}, file string, line int) ErrorFunc {
	switch v.(type) {
	case error:
		var e ErrorFunc = func() *Error { return &Error{time.Now(), file, line, v.(error).Error()} }
		return e
	case string:
		var e ErrorFunc = func() *Error { return &Error{time.Now(), file, line, v.(string)} }
		return e
	default:
		var e ErrorFunc = func() *Error { return &Error{time.Now(), file, line, fmt.Sprintf("%v", v)} }
		return e
	}
}

func NewStackWithError(deep int, v interface{}) ErrorFunc {
	deep = 10 + deep*2
	var file, line = caller.Stack(deep)
	return newErrorWithFileAndLine(v, file, line)
}

func Parse(err interface{}) ErrorFunc {
	switch err.(type) {
	case ErrorFunc:
		return err.(ErrorFunc)
	case *Error:
		return func() *Error { return err.(*Error) }
	default:
		file, line := caller.Caller(2)
		return func() *Error {
			return &Error{Time: time.Now(), File: file, Line: line, Message: fmt.Sprintf("%v", err)}
		}
	}
}

func IsNil(i interface{}) bool {
	if i == nil {
		return true
	}
	vi := reflect.ValueOf(i)

	switch vi.Kind() {
	case reflect.UnsafePointer:
		return vi.IsNil()
	case reflect.Ptr:
		return vi.IsNil()
	case reflect.Chan:
		return vi.IsNil()
	case reflect.Func:
		return vi.IsNil()
	case reflect.Interface:
		return vi.IsNil()
	case reflect.Map:
		return vi.IsNil()
	case reflect.Slice:
		return vi.IsNil()
	}

	return false
}
