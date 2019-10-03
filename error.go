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
	"errors"
	"fmt"
	"runtime"
	"time"
)

type Error struct {
	Time  time.Time
	File  string
	Line  int
	Error error
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
			return &Error{time.Now(), file, line, err.(error)}
		}
	case string:
		return func() *Error {
			return &Error{time.Now(), file, line, errors.New(err.(string))}
		}
	default:
		return func() *Error {
			return &Error{time.Now(), file, line, fmt.Errorf("%s", err)}
		}
	}

}
