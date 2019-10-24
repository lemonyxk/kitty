package logger

import (
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/gookit/color"

	"github.com/Lemo-yxk/lemo"
)

const DEBUG int = 1
const LOG int = 2

var debug = false
var log = false

func SetDebug(v bool) {
	debug = v
}

func SetLog(v bool) {
	log = v
}

func SetFlag(flag int) {
	switch flag {
	case 3:
		debug = true
		log = true
	case 2:
		log = true
	case 1:
		debug = true
	case 0:
		debug = false
		log = false
	default:
		return
	}
}

type Logger struct {
	debugHook func(t time.Time, file string, line int, v ...interface{})
	errorHook func(err *lemo.Error)
	logHook   func(status string, t time.Time, file string, line int, v ...interface{})
}

var logger *Logger

func init() {

	logger = new(Logger)

	SetDebug(true)
	SetLog(false)

	SetDebugHook(func(t time.Time, file string, line int, v ...interface{}) {
		var date = time.Now().Format("2006-01-02 15:04:05")
		color.Blue.Print(date + " " + file + ":" + strconv.Itoa(line) + " ")
		color.Blue.Println(v...)
	})

	SetErrorHook(func(err *lemo.Error) {
		var date = err.Time.Format("2006-01-02 15:04:05")
		color.Red.Println(date + " " + err.File + ":" + strconv.Itoa(err.Line) + " " + err.Error.Error())
	})

	SetLogHook(func(status string, t time.Time, file string, line int, v ...interface{}) {})
}

func Println(v ...interface{}) {
	color.Println(v...)
}

func SetDebugHook(fn func(t time.Time, file string, line int, v ...interface{})) {
	logger.debugHook = fn
}

func SetErrorHook(fn func(err *lemo.Error)) {
	logger.errorHook = fn
}

func SetLogHook(fn func(status string, t time.Time, file string, line int, v ...interface{})) {
	logger.logHook = fn
}

func Console(v ...interface{}) {

	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return
	}

	var t = time.Now()

	if debug {
		logger.debugHook(t, file, line, v...)
	}

	if log {
		logger.logHook("CONSOLE", t, file, line, v...)
	}
}

func Error(err interface{}) {

	switch err.(type) {
	case func() *lemo.Error:
		var res = err.(func() *lemo.Error)()

		if debug {
			logger.errorHook(res)
		}

		if log {
			logger.logHook("ERROR", res.Time, res.File, res.Line, res.Error)
		}

	case *lemo.Error:

		var res = err.(*lemo.Error)

		if debug {
			logger.errorHook(res)
		}

		if log {
			logger.logHook("ERROR", res.Time, res.File, res.Line, res.Error)
		}
	default:

		_, file, line, ok := runtime.Caller(1)
		if !ok {
			return
		}

		var t = time.Now()

		if debug {
			logger.errorHook(&lemo.Error{Time: t, File: file, Line: line, Error: fmt.Errorf("%v", err)})
		}

		if log {
			logger.logHook("ERROR", t, file, line, err)
		}

	}

}
