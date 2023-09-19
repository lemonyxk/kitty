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
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/lemonyxk/caller"
	"github.com/modern-go/reflect2"
)

var space = strings.Repeat(" ", 4) + "at "
var withStack = true

func WithStack(b bool) {
	withStack = b
}

type Error struct {
	message string
	err     error
	stack   []caller.Info
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
				var str = space + filepath.Base(f.Func) + " in " + f.File + ":" + strconv.Itoa(f.Line)
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

func New(text any) error {
	if reflect2.IsNil(text) {
		return nil
	}

	switch text.(type) {
	case *Error:
		return text.(*Error)
	case error:
		var r = &Error{message: text.(error).Error()}
		if withStack {
			r.stack = caller.Deeps(2)
		}
		return r
	default:
		var r = &Error{message: fmt.Sprintf("%+v", text)}
		if withStack {
			r.stack = caller.Deeps(2)
		}
		return r
	}
}

func Errorf(f string, args ...any) error {
	var r = &Error{message: fmt.Sprintf(f, args...)}
	if withStack {
		r.stack = caller.Deeps(2)
	}
	return r
}

func Wrap(err error, text any) error {
	if err == nil {
		return nil
	}

	var r = &Error{
		message: fmt.Sprintf("%+v", text) + ": " + err.Error(),
		err:     err,
	}

	if e, ok := err.(*Error); ok {
		r.stack = e.stack
		return r
	}

	if withStack {
		r.stack = caller.Deeps(2)
	}

	return r
}

func Wrapf(err error, f string, args ...any) error {
	if err == nil {
		return nil
	}

	var r = &Error{
		message: fmt.Sprintf(f, args...) + ": " + err.Error(),
		err:     err,
	}

	if e, ok := err.(*Error); ok {
		r.stack = e.stack
		return r
	}

	if withStack {
		r.stack = caller.Deeps(2)
	}

	return r
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}
