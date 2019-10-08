package logger

import (
	"bytes"
	"fmt"
	"runtime"
	"time"

	"github.com/gookit/color"

	"github.com/Lemo-yxk/lemo"
)

type statementServerMessage struct {
	Type string `json:"type"`
	From string `json:"from"`
	Host string `json:"host"`
	File string `json:"file"`
	Line int    `json:"line"`
	Msg  string `json:"msg"`
	Time string `json:"time"`
}

var debug = true
var log = false

func SetDebug(flag bool) {
	debug = flag
}

func SetLog(flag bool) {
	log = flag
}

type Logger struct{}

var logger *Logger

func init() {
	logger = new(Logger)
}

func (logger *Logger) debugHook(t time.Time, file string, line int, v ...interface{}) {
	var date = time.Now().Format("2006-01-02 15:04:05")

	var buf bytes.Buffer

	for index, value := range v {
		buf.WriteString(fmt.Sprint(value))
		if index != len(v)-1 {
			buf.WriteString(" ")
		}
	}

	color.Blue.Println(fmt.Sprintf("%s %s:%d %s", date, file, line, buf.String()))
}

func (logger *Logger) errorHook(err *lemo.Error) {
	var date = err.Time.Format("2006-01-02 15:04:05")
	color.Red.Println(date, fmt.Sprintf("%s:%d", err.File, err.Line), err.Error)
}

func (logger *Logger) logHook(t time.Time, file string, line int, v ...interface{}) {}

func Log(v ...interface{}) {

	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return
	}

	var t = time.Now()

	if debug {
		logger.debugHook(t, file, line, v...)
	}

	if log {
		logger.logHook(t, file, line, v...)
	}
}

func Err(err interface{}) {

	switch err.(type) {
	case func() *lemo.Error:
		var res = err.(func() *lemo.Error)()

		if debug {
			logger.errorHook(res)
		}

		if log {
			logger.logHook(res.Time, res.File, res.Line, res.Error)
		}

	case *lemo.Error:

		var res = err.(*lemo.Error)

		if debug {
			logger.errorHook(res)
		}

		if log {
			logger.logHook(res.Time, res.File, res.Line, res.Error)
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
			logger.logHook(t, file, line, err)
		}

	}

}
