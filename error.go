/**
* @program: lemo
*
* @description:
*
* @author: Mr.Wang
*
* @create: 2019-09-25 20:37
**/

package lemo

import (
	"errors"
	"fmt"
	"runtime"
)

type Error struct {
	File  string
	Line  int
	Error error
}

func NewError(err interface{}) func() *Error {

	if err == nil {
		return nil
	}

	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return nil
	}

	switch err.(type) {
	case error:
		return func() *Error {
			return &Error{file, line, err.(error)}
		}
	case string:
		return func() *Error {
			return &Error{file, line, errors.New(err.(string))}
		}
	default:
		return func() *Error {
			return &Error{file, line, fmt.Errorf("%s", err)}
		}
	}

}
