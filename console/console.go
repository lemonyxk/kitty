package console

import (
	"fmt"
	"os"
	"time"

	"github.com/Lemo-yxk/lemo/caller"
	"github.com/Lemo-yxk/lemo/exception"
	"github.com/Lemo-yxk/lemo/utils"
)

var hook = false
var output = true

func Hook(v bool) {
	hook = v
}

func OutPut(v bool) {
	output = v
}

type Logger struct {
	debugHook   func(t time.Time, file string, line int, v ...interface{})
	logHook     func(t time.Time, file string, line int, v ...interface{})
	warningHook func(t time.Time, file string, line int, v ...interface{})
	errorHook   func(err *exception.Error)
	hook        func(status string, t time.Time, file string, line int, v ...interface{})
}

var logger *Logger

func init() {

	logger = new(Logger)

	SetDebugHook(func(t time.Time, file string, line int, v ...interface{}) {
		FgBlue.Printf("DEB %s %s:%d %s \n", time.Now().Format("2006-01-02 15:04:05"), file, line, utils.String.JoinInterface(v, " "))
	})

	SetLogHook(func(t time.Time, file string, line int, v ...interface{}) {
		Bold.Printf("LOG %s %s:%d %s \n", time.Now().Format("2006-01-02 15:04:05"), file, line, utils.String.JoinInterface(v, " "))
	})

	SetWarningHook(func(t time.Time, file string, line int, v ...interface{}) {
		FgYellow.Printf("WAR %s %s:%d %s \n", time.Now().Format("2006-01-02 15:04:05"), file, line, utils.String.JoinInterface(v, " "))
	})

	SetErrorHook(func(err *exception.Error) {
		FgRed.Printf("ERR %s %s:%d %s \n", err.Time.Format("2006-01-02 15:04:05"), err.File, err.Line, err.Message)
	})

	SetHook(nil)
}

func Exit(v interface{}) {
	Error(v)
	os.Exit(0)
}

func SetDebugHook(fn func(t time.Time, file string, line int, v ...interface{})) {
	logger.debugHook = fn
}

func SetLogHook(fn func(t time.Time, file string, line int, v ...interface{})) {
	logger.logHook = fn
}

func SetWarningHook(fn func(t time.Time, file string, line int, v ...interface{})) {
	logger.warningHook = fn
}

func SetErrorHook(fn func(err *exception.Error)) {
	logger.errorHook = fn
}

func SetHook(fn func(status string, t time.Time, file string, line int, v ...interface{})) {
	logger.hook = fn
}

func Println(v ...interface{}) {
	fmt.Println(v...)
}

func Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

func Warning(v ...interface{}) {
	file, line := caller.Caller(1)

	var t = time.Now()

	if output && logger.warningHook != nil {
		logger.warningHook(t, file, line, v...)
	}

	if hook && logger.hook != nil {
		logger.hook("WAR", t, file, line, v...)
	}
}

func Debug(v ...interface{}) {
	file, line := caller.Caller(1)

	var t = time.Now()

	if output && logger.debugHook != nil {
		logger.debugHook(t, file, line, v...)
	}

	if hook && logger.hook != nil {
		logger.hook("DEB", t, file, line, v...)
	}
}

func Log(v ...interface{}) {
	file, line := caller.Caller(1)

	var t = time.Now()

	if output && logger.logHook != nil {
		logger.logHook(t, file, line, v...)
	}

	if hook && logger.hook != nil {
		logger.hook("LOG", t, file, line, v...)
	}
}

func Customize(color Color, tp string, format string, v ...interface{}) {
	file, line := caller.Caller(1)

	var t = time.Now()

	color.Printf(format, v...)

	if hook && logger.hook != nil {
		logger.hook(tp, t, file, line, v...)
	}
}

func Error(err interface{}) {

	switch err.(type) {
	case exception.ErrorFunc:

		var res = err.(exception.ErrorFunc)

		if res == nil {
			printDefault(res)
			return
		}

		var r = res()

		if r == nil {
			printDefault(r)
			return
		}

		printError(r)
	case *exception.Error:
		var r = err.(*exception.Error)

		if r == nil {
			printDefault(r)
			return
		}

		printError(r)
	default:
		printDefault(err)
	}
}

func printError(err *exception.Error) {
	if output && logger.errorHook != nil {
		logger.errorHook(err)
	}

	if hook && logger.hook != nil {
		logger.hook("ERR", err.Time, err.File, err.Line, err.Message)
	}
}

func printDefault(err interface{}) {
	file, line := caller.Caller(2)

	var t = time.Now()

	if output && logger.errorHook != nil {
		logger.errorHook(&exception.Error{Time: t, File: file, Line: line, Message: fmt.Sprintf("%v", err)})
	}

	if hook && logger.hook != nil {
		logger.hook("ERR", t, file, line, err)
	}
}
