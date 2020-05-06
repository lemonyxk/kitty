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
	errorHook   func(err exception.Error)
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

	SetErrorHook(func(err exception.Error) {
		FgRed.Printf("ERR %s %s:%d %s \n", err.Time().Format("2006-01-02 15:04:05"), err.File(), err.Line(), err.Error())
	})

	SetHook(nil)
}

func Exit(v interface{}) {
	errorWithStack(v, 3)
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

func SetErrorHook(fn func(err exception.Error)) {
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

func OneLine(format string, v ...interface{}) {
	fmt.Printf("\r"+format, v...)
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

func Customize(color Color, prefix string, format string, v ...interface{}) {
	file, line := caller.Caller(1)

	var t = time.Now()

	if output {
		color.Printf(format, v...)
	}

	if hook && logger.hook != nil {
		logger.hook(prefix, t, file, line, v...)
	}
}

func Assert(v ...interface{}) {
	if len(v) == 0 {
		return
	}
	if exception.IsNil(v[len(v)-1]) {
		return
	}
	errorWithStack(v[len(v)-1], 3)
}

func Error(err interface{}) {
	errorWithStack(err, 3)
}

func errorWithStack(err interface{}, deep int) {

	switch err.(type) {
	case exception.Error:
		printError(err.(exception.Error))
	default:
		printDefault(err, deep)
	}
}

func printError(err exception.Error) {
	if output && logger.errorHook != nil {
		logger.errorHook(err)
	}

	if hook && logger.hook != nil {
		logger.hook("ERR", err.Time(), err.File(), err.Line(), err.Error())
	}
}

func printDefault(err interface{}, deep int) {

	var res = exception.NewErrorFromDeep(err, deep+1)

	if output && logger.errorHook != nil {
		logger.errorHook(res)
	}

	if hook && logger.hook != nil {
		logger.hook("ERR", res.Time(), res.File(), res.Line(), res.Error())
	}
}
