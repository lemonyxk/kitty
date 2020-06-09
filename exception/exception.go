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
	"unsafe"

	"github.com/Lemo-yxk/lemo/caller"
)

type exception struct {
	time  time.Time
	file  string
	line  int
	error string
}

func (e exception) Time() time.Time {
	return e.time
}

func (e exception) File() string {
	return e.file
}

func (e exception) Line() int {
	return e.line
}

func (e exception) Error() string {
	return e.error
}

func (e exception) String() string {
	return fmt.Sprintf("ERR %s %s:%d %s", e.time.Format("2006-01-02 15:04:05"), e.file, e.line, e.error)
}

func NewException(time time.Time, file string, line int, err string) Error {
	return exception{time: time, file: file, line: line, error: err}
}

type Error interface {
	Time() time.Time
	File() string
	Line() int
	Error() string
	String() string
}

type catchFunc func(err Error)

type finallyFunc func(err Error)

type catch struct {
	fn func(catchFunc) *finally
}

func (c *catch) Catch(fn catchFunc) *finally {
	return c.fn(fn)
}

type finally struct {
	fn  func(finallyFunc)
	err Error
}

func (f *finally) Finally(fn finallyFunc) Error {
	f.fn(fn)
	return f.err
}

func (f *finally) Error() Error {
	return f.err
}

func Try(fn func()) (c *catch) {

	defer func() {
		if err := recover(); err != nil {
			var d = 1
			var e = fmt.Errorf("%v", err)
			if strings.HasPrefix(e.Error(), "#exception#") {
				d = 2
			}
			var stacks = newStackErrorFromDeep(strings.Replace(e.Error(), "#exception#", "", 1), d)
			c = &catch{fn: func(f catchFunc) *finally {
				f(stacks)
				return &finally{
					err: stacks,
					fn:  func(ff finallyFunc) { ff(stacks) },
				}
			}}
		}
	}()

	fn()

	return &catch{fn: func(f catchFunc) *finally {
		return &finally{
			fn: func(ff finallyFunc) { ff(nil) },
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

func New(v ...interface{}) Error {
	if len(v) == 0 {
		return nil
	}
	if IsNil(v[len(v)-1]) {
		return nil
	}
	return newErrorFromDeep(v[len(v)-1], 2)
}

func NewMany(v ...interface{}) Error {
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

	return newErrorFromDeep(str[0:len(str)-1], 2)
}

func NewFormat(format string, v ...interface{}) Error {
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

func newErrorFromDeep(v interface{}, deep int) Error {
	file, line := caller.Caller(deep)
	return newErrorWithFileAndLine(v, file, line)
}

func newStackErrorFromDeep(v interface{}, deep int) Error {
	deep = 10 + deep*2
	var file, line = caller.Stack(deep)
	return newErrorWithFileAndLine(v, file, line)
}

func newErrorWithFileAndLine(v interface{}, file string, line int) Error {
	var err string

	switch v.(type) {
	case error:
		err = v.(error).Error()
	case string:
		err = v.(string)
	case Error:
		return v.(Error)
	default:
		err = fmt.Sprintf("%v", v)
	}

	return NewException(time.Now(), file, line, err)
}

func IsNil(i interface{}) bool {
	if i == nil {
		return true
	}
	return (*eFace)(unsafe.Pointer(&i)).data == nil
}

type eFace struct {
	_type unsafe.Pointer
	data  unsafe.Pointer
}
