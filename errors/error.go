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
	"bytes"
	"errors"
	"fmt"
	"github.com/lemonyxk/kitty/kitty"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/lemonyxk/caller"
)

var space = strings.Repeat(" ", 4) + "at "
var withStack = true

func WithStack(b bool) {
	withStack = b
}

type Error struct {
	errs  []error
	stack []caller.Info
	buf   *bytes.Buffer
}

func (e *Error) Error() string {
	if len(e.errs) == 0 {
		return ""
	}

	if e.buf == nil {
		e.buf = new(bytes.Buffer)
		for i := len(e.errs) - 1; i >= 0; i-- {
			if i != len(e.errs)-1 {
				_, _ = io.WriteString(e.buf, ": ")
			}
			_, _ = io.WriteString(e.buf, e.errs[i].Error())
		}
	}
	return e.buf.String()
}

func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = io.WriteString(s, e.Error())
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
		_, _ = io.WriteString(s, e.Error())
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", e.Error())
	}
}

func (e *Error) Unwrap() error {
	if len(e.errs) == 0 {
		return nil
	}
	var err = e.errs[len(e.errs)-1]
	e.errs = e.errs[:len(e.errs)-1]
	return err
}

func New(text any) error {

	if kitty.IsNil(text) {
		return nil
	}

	switch text.(type) {
	case *Error:
		return text.(*Error)
	case error:
		var r = &Error{errs: []error{text.(error)}}
		if withStack {
			r.stack = caller.Deeps(2)
		}
		return r
	case string:
		var r = &Error{errs: []error{errors.New(text.(string))}}
		if withStack {
			r.stack = caller.Deeps(2)
		}
		return r
	default:
		var r = &Error{errs: []error{fmt.Errorf("%+v", text)}}
		if withStack {
			r.stack = caller.Deeps(2)
		}
		return r
	}
}

func Errorf(f string, args ...any) error {
	var r = &Error{errs: []error{fmt.Errorf(f, args...)}}
	if withStack {
		r.stack = caller.Deeps(2)
	}
	return r
}

func Wrap(err error, text any) error {
	if kitty.IsNil(text) {
		return err
	}

	var r = &Error{}

	if err != nil {
		r.errs = append(r.errs, err)
	}

	switch text.(type) {
	case *Error:
		r.errs = append(r.errs, text.(*Error))
	case error:
		r.errs = append(r.errs, text.(error))
	case string:
		r.errs = append(r.errs, errors.New(text.(string)))
	default:
		r.errs = append(r.errs, fmt.Errorf("%+v", text))
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
	if len(args) == 0 {
		return err
	}

	var r = &Error{}

	if err != nil {
		r.errs = append(r.errs, err)
	}

	r.errs = append(r.errs, fmt.Errorf(f, args...))

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
	if e, ok := err.(*Error); ok {
		if e1, ok1 := target.(*Error); ok1 {
			for i := 0; i < len(e.errs); i++ {
				for j := 0; j < len(e1.errs); j++ {
					if errors.Is(e.errs[i], e1.errs[j]) {
						return true
					}
				}
			}
			return false
		}
		for i := 0; i < len(e.errs); i++ {
			if errors.Is(e.errs[i], target) {
				return true
			}
		}
		return false
	} else {
		if e1, ok1 := target.(*Error); ok1 {
			for i := 0; i < len(e1.errs); i++ {
				if errors.Is(err, e1.errs[i]) {
					return true
				}
			}
			return false
		}
		return errors.Is(err, target)
	}
}

func Unwrap(err error) error {
	if e, ok := err.(*Error); ok {
		return e.Unwrap()
	}
	return errors.Unwrap(err)
}
