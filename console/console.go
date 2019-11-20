package console

import (
	"fmt"
	"github.com/Lemo-yxk/lemo/exception"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/gookit/color"
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
	errorHook func(err *exception.Error)
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

	SetErrorHook(func(err *exception.Error) {
		var date = err.Time.Format("2006-01-02 15:04:05")
		color.Red.Println(date + " " + err.File + ":" + strconv.Itoa(err.Line) + " " + err.Message)
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
	color.Println(v...)
}

func Log(v ...interface{}) {

	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return
	}

	var t = time.Now()

	if debug {
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

		if debug {
			logger.errorHook(res)
		}

		if log && logger.logHook != nil {
			logger.logHook("ERROR", res.Time, res.File, res.Line, res.Message)
		}

	case *exception.Error:

		var res = err.(*exception.Error)

		if debug {
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

		if debug {
			logger.errorHook(&exception.Error{Time: t, File: file, Line: line, Message: fmt.Sprintf("%s", err)})
		}

		if log && logger.logHook != nil {
			logger.logHook("ERROR", t, file, line, err)
		}

	}

}
