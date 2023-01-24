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

// var pwd, _ = os.Getwd()

type info struct {
	file     string
	line     int
	funcName string
}

type errString struct {
	message string
	err     error
	stack   []info
}

func (e *errString) Error() string {
	return e.message
}

func (e *errString) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = io.WriteString(s, e.message+"\n")
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

func New(text string) error {
	return &errString{message: text}
}

func NewWithStack(text string) error {
	return &errString{message: text, stack: stack(2)}
}

func stack(deep int) []info {
	var res []info
	for skip := deep; true; skip++ {
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

		codePath = codePath[index:]
		prevFunc := runtime.FuncForPC(pc).Name()
		res = append(res, info{
			file:     codePath,
			line:     codeLine,
			funcName: prevFunc,
		})
	}
	return res
}

func WithStack(err error) error {
	if err == nil {
		return nil
	}
	return &errString{message: err.Error(), stack: stack(2)}
}

func Wrap(err error, text string) error {
	if err == nil {
		return nil
	}

	var r = &errString{
		message: text + ": " + err.Error(),
		err:     err,
	}

	if e, ok := err.(*errString); ok {
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

func (e *errString) Unwrap() error {
	return e.err
}
