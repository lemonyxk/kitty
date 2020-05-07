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

type errors struct {
	time  time.Time
	file  string
	line  int
	error string
}

func (err errors) Time() time.Time {
	return err.time
}

func (err errors) File() string {
	return err.file
}

func (err errors) Line() int {
	return err.line
}

func (err errors) Error() string {
	return err.error
}

func (err errors) String() string {
	return fmt.Sprintf("ERR %s %s:%d %s", err.time.Format("2006-01-02 15:04:05"), err.file, err.line, err.error)
}

type Error interface {
	Time() time.Time
	File() string
	Line() int
	Error() string
	String() string
}

type Catch func(Error)

type Finally func(Error)

type catch struct {
	fn func(Catch) *finally
}

func (c *catch) Catch(fn Catch) *finally {
	return c.fn(fn)
}

type finally struct {
	fn  func(Finally)
	err Error
}

func (f *finally) Finally(fn Finally) Error {
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
			var stacks = NewStackErrorFromDeep(strings.Replace(e.Error(), "#exception#", "", 1), d)
			c = &catch{fn: func(f Catch) *finally {
				f(stacks)
				return &finally{
					err: stacks,
					fn:  func(ff Finally) { ff(stacks) },
				}
			}}
		}
	}()

	fn()

	return &catch{fn: func(f Catch) *finally {
		return &finally{
			fn: func(ff Finally) { ff(nil) },
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

func Inspect(v ...interface{}) Error {
	if len(v) == 0 {
		return nil
	}
	if IsNil(v[len(v)-1]) {
		return nil
	}
	return NewErrorFromDeep(v[len(v)-1], 2)
}

func New(v ...interface{}) Error {
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
		return NewErrorFromDeep(v[0], 2)
	}

	var str = fmt.Sprintln(v...)
	return NewErrorFromDeep(str[:len(str)-1], 2)
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
	return NewErrorFromDeep(str, 2)
}

func NewErrorFromDeep(v interface{}, deep int) Error {
	file, line := caller.Caller(deep)
	return newErrorWithFileAndLine(v, file, line)
}

func NewStackErrorFromDeep(v interface{}, deep int) Error {
	deep = 10 + deep*2
	var file, line = caller.Stack(deep)
	return newErrorWithFileAndLine(v, file, line)
}

func newErrorWithFileAndLine(v interface{}, file string, line int) Error {
	switch v.(type) {
	case error:
		return errors{time.Now(), file, line, v.(error).Error()}
	case string:
		return errors{time.Now(), file, line, v.(string)}
	case Error:
		return v.(Error)
	default:
		return errors{time.Now(), file, line, fmt.Sprintf("%v", v)}
	}
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
