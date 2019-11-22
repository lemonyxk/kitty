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
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

type Error struct {
	Time    time.Time
	File    string
	Line    int
	Message string
}

var Empty = func() func() *Error {
	return func() *Error {
		return nil
	}
}

func (err *Error) Error() string {
	return err.String()
}

func (err *Error) String() string {
	return err.Time.Format("2006-01-02 15:04:05") + " " + err.File + ":" + strconv.Itoa(err.Line) + " " + err.Message
}

func Panic(err interface{}) {
	panic(string(debug.Stack()))
}

func New(err ...interface{}) func() *Error {
	if len(err) == 0 {
		return nil
	}

	if len(err) == 1 {
		if err[0] == nil {
			return nil
		}
		return newErrorFromDeep(err[0], 2)
	}

	if len(err) > 1 {
		if format, ok := err[0].(string); ok && len(format) > 1 && strings.Index(format, "%") != -1 {
			return newErrorFromDeep(fmt.Errorf(format, err[1:]...), 2)
		}
	}

	return newErrorFromDeep(fmt.Sprintln(err...), 2)
}

func newErrorFromDeep(err interface{}, deep int) func() *Error {

	if err == nil {
		return nil
	}

	_, file, line, ok := runtime.Caller(deep)
	if !ok {
		return nil
	}

	switch err.(type) {
	case error:
		return func() *Error {
			return &Error{time.Now(), file, line, err.(error).Error()}
		}
	case string:
		return func() *Error {
			return &Error{time.Now(), file, line, err.(string)}
		}
	default:
		return func() *Error {
			return &Error{time.Now(), file, line, fmt.Sprintf("%s", err)}
		}
	}

}
