/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2022-05-23 04:45
**/

package errors

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

var pwd, _ = os.Getwd()
var goRoot, _ = os.LookupEnv("GOROOT")

type info struct {
	file     string
	line     int
	funcName string
}

type Error struct {
	message string
	err     error
	stack   []info
}

func (e *Error) Error() string {
	return e.message
}

func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = io.WriteString(s, e.message)
			if len(e.stack) != 0 {
				_, _ = io.WriteString(s, "\n")
			}
			for i, f := range e.stack {
				var str = strings.Repeat(" ", 4) + "at " + filepath.Base(f.funcName) + " in " + f.file + ":" + strconv.Itoa(f.line)
				if i != len(e.stack)-1 {
					str = str + "\n"
				}
				_, _ = io.WriteString(s, str)
			}
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(s, e.message)
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", e.message)
	}
}

func (e *Error) Unwrap() error {
	return e.err
}

func (e *Error) Stack() error {
	e.stack = stack(2)
	return e
}

func New(text string) *Error {
	return &Error{message: text}
}

func Errorf(f string, args ...any) *Error {
	return &Error{message: fmt.Sprintf(f, args...)}
}

func NewWithStack(text string) error {
	return &Error{message: text, stack: stack(2)}
}

func WithStack(err error) error {
	if err == nil {
		return nil
	}
	return &Error{message: err.Error(), stack: stack(2)}
}

func Wrap(err error, text string) error {
	if err == nil {
		return nil
	}

	var r = &Error{
		message: text + ": " + err.Error(),
		err:     err,
	}

	if e, ok := err.(*Error); ok {
		r.stack = e.stack
		return r
	}

	return r
}

func Is(err, target error) bool {
	if target == nil {
		return err == target
	}

	isComparable := reflect.TypeOf(target).Comparable()
	for {
		if isComparable && err == target {
			return true
		}
		if x, ok := err.(interface{ Is(error) bool }); ok && x.Is(target) {
			return true
		}
		if err = Unwrap(err); err == nil {
			return false
		}
	}
}

func Unwrap(err error) error {
	u, ok := err.(interface {
		Unwrap() error
	})
	if !ok {
		return nil
	}
	return u.Unwrap()
}

func stack(deep int) []info {
	var res []info
	for skip := 2; true; skip++ {

		pc, codePath, codeLine, ok := runtime.Caller(skip)
		if !ok {
			break
		}

		var index = 0
		var count = 0
		for i := len(codePath) - 1; i >= 0; i-- {
			if codePath[i] == os.PathSeparator {
				count++
			}
			if count == 3 {
				index = i + 1
				break
			}
		}

		if codePath[:len(pwd)] == pwd {
			codePath = codePath[len(pwd)+1:]
		} else if codePath[:len(goRoot)] == goRoot {
			// codePath = codePath[len(goRoot)+1:]
		} else {
			codePath = "@" + codePath[index:]
		}

		prevFunc := runtime.FuncForPC(pc).Name()
		res = append(res, info{
			file:     codePath,
			line:     codeLine,
			funcName: prevFunc,
		})
	}
	return res
}
