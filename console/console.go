package console

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/Lemo-yxk/lemo/exception"
)

type Content []interface{}

var (
	debug      = true
	e          = true
	log        = false
	debugColor = FgBlue
	errorColor = FgRed
)

func SetDebug(v bool) {
	debug = v
}

func SetError(v bool) {
	e = v
}

func SetLog(v bool) {
	log = v
}

func SetDebugColor(color Color) {
	debugColor = color
}

func SetErrorColor(color Color) {
	errorColor = color
}

type Logger struct {
	debugHook func(t time.Time, file string, line int, v ...interface{})
	errorHook func(err *exception.Error)
	logHook   func(status string, t time.Time, file string, line int, v ...interface{})
}

var logger *Logger

func init() {
	logger = new(Logger)

	SetDebug(true)
	SetLog(false)

	SetDebugHook(func(t time.Time, file string, line int, v ...interface{}) {
		debugColor.Println(append(Content{time.Now().Format("2006-01-02 15:04:05") + " " + file + ":" + strconv.Itoa(line) + " "}, v...)...)
	})

	SetErrorHook(func(err *exception.Error) {
		errorColor.Println(err.Time.Format("2006-01-02 15:04:05") + " " + err.File + ":" + strconv.Itoa(err.Line) + " " + err.Message)
	})

	SetLogHook(nil)
}

func Exit(v interface{}) {
	Error(v)
	os.Exit(0)
}

func SetDebugHook(fn func(t time.Time, file string, line int, v ...interface{})) {
	logger.debugHook = fn
}

func SetErrorHook(fn func(err *exception.Error)) {
	logger.errorHook = fn
}

func SetLogHook(fn func(status string, t time.Time, file string, line int, v ...interface{})) {
	logger.logHook = fn
}

func Println(v ...interface{}) {
	fmt.Println(v...)
}

func Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

func Log(v ...interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return
	}

	var t = time.Now()

	if debug && logger.debugHook != nil {
		logger.debugHook(t, file, line, v...)
	}

	if log && logger.logHook != nil {
		logger.logHook("CONSOLE", t, file, line, v...)
	}
}

func Error(err interface{}) {

	switch err.(type) {
	case func() *exception.Error:
		var res = err.(func() *exception.Error)()

		if res == nil {
			return
		}

		if e && logger.errorHook != nil {
			logger.errorHook(res)
		}

		if log && logger.logHook != nil {
			logger.logHook("ERROR", res.Time, res.File, res.Line, res.Message)
		}
	case *exception.Error:
		var res = err.(*exception.Error)

		if res == nil {
			return
		}

		if e && logger.errorHook != nil {
			logger.errorHook(res)
		}

		if log && logger.logHook != nil {
			logger.logHook("ERROR", res.Time, res.File, res.Line, res.Message)
		}
	default:
		_, file, line, ok := runtime.Caller(1)
		if !ok {
			return
		}

		var t = time.Now()

		if e && logger.errorHook != nil {
			logger.errorHook(&exception.Error{Time: t, File: file, Line: line, Message: fmt.Sprintf("%s", err)})
		}

		if log && logger.logHook != nil {
			logger.logHook("ERROR", t, file, line, err)
		}
	}
}
