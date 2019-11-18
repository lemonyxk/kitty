/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-09-25 20:37
**/

package lemo

import (
	"fmt"
	"runtime"
	"strconv"
	"time"
)

type Error struct {
	Time    time.Time
	File    string
	Line    int
	Message string
}

func (err *Error) Error() string {
	return err.String()
}

func (err *Error) String() string {
	return err.Time.Format("2006-01-02 15:04:05") + " " + err.File + ":" + strconv.Itoa(err.Line) + " " + err.Message
}

func NewError(err interface{}) func() *Error {
	return NewErrorFromDeep(err, 2)
}

func NewErrorFromDeep(err interface{}, deep int) func() *Error {

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
