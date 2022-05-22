/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2022-05-23 04:45
**/

package errors

import (
	"fmt"
	"io"
	"runtime"
	"strconv"
)

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
			for _, f := range e.stack {
				var str = f.funcName + "\n\t" + f.file + ":" + strconv.Itoa(f.line) + "\n"
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
	var e = &errString{message: text}
	e.stack = stack(2)
	return e
}

func stack(deep int) []info {
	var res []info
	for skip := deep; true; skip++ {
		pc, codePath, codeLine, ok := runtime.Caller(skip)
		if !ok {
			break
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

func WithStack(err error) error {
	if err == nil {
		return nil
	}
	var e = &errString{message: err.Error()}
	e.stack = stack(2)
	return e
}

func Wrap(err error, text string) error {
	if err == nil {
		return nil
	}
	return &errString{
		message: text + ": " + err.Error(),
		err:     err,
		stack:   stack(2),
	}
}

func (e *errString) Unwrap() error {
	return e.err
}
