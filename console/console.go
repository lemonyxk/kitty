package console

import (
	"fmt"
	"os"
	"time"

	"github.com/Lemo-yxk/lemo/caller"
	"github.com/Lemo-yxk/lemo/exception"
	"github.com/Lemo-yxk/lemo/utils"
)

type Logger interface {
	Errorf(string, ...interface{})
	Warningf(string, ...interface{})
	Infof(string, ...interface{})
	Debugf(string, ...interface{})
}

var wr *writer

type writer struct {
	logger Logger
}

func SetLogger(logger Logger) {
	wr.logger = logger
}

var DefaultLogger *defaultLogger

type defaultLogger struct {
	Hook func(status string, t time.Time, file string, line int, v ...interface{})
}

func (log *defaultLogger) Errorf(format string, args ...interface{}) {

	var err interface{}

	if len(args) == 0 {
		err = ""
	}

	err = args[len(args)-1]

	var file string
	var line int
	var t time.Time
	var msg string

	switch err.(type) {
	case exception.Error:
		file, line = err.(exception.Error).File(), err.(exception.Error).Line()
		msg = err.(exception.Error).Error()
		t = err.(exception.Error).Time()
	default:
		file, line = caller.Caller(2)
		msg = fmt.Sprintf("%s", err)
		t = time.Now()
	}

	FgRed.Printf(format, t.Format("2006-01-02 15:04:05"), file, line, msg)

	if log.Hook != nil {
		log.Hook("ERR", t, file, line, err)
	}
}

func (log *defaultLogger) Warningf(format string, args ...interface{}) {
	file, line := caller.Caller(2)
	msg := utils.String.JoinInterface(args, " ")
	t := time.Now()

	FgYellow.Printf(format, t.Format("2006-01-02 15:04:05"), file, line, msg)

	if log.Hook != nil {
		log.Hook("WAR", t, file, line, args...)
	}
}

func (log *defaultLogger) Infof(format string, args ...interface{}) {
	file, line := caller.Caller(2)
	msg := utils.String.JoinInterface(args, " ")
	t := time.Now()

	Bold.Printf(format, t.Format("2006-01-02 15:04:05"), file, line, msg)

	if log.Hook != nil {
		log.Hook("LOG", t, file, line, args...)
	}
}

func (log *defaultLogger) Debugf(format string, args ...interface{}) {
	file, line := caller.Caller(2)
	msg := utils.String.JoinInterface(args, " ")
	t := time.Now()

	FgBlue.Printf(format, t.Format("2006-01-02 15:04:05"), file, line, msg)

	if log.Hook != nil {
		log.Hook("DEB", t, file, line, args...)
	}
}

func init() {
	wr = new(writer)
	DefaultLogger = new(defaultLogger)
	SetLogger(DefaultLogger)
}

func Exit(v interface{}) {
	wr.logger.Errorf("ERR %s %s:%d %s \n", v)
	os.Exit(0)
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

func Log(v ...interface{}) {
	wr.logger.Infof("LOG %s %s:%d %s \n", v...)
}

func Debug(v ...interface{}) {
	wr.logger.Debugf("DEB %s %s:%d %s \n", v...)
}

func Warning(v ...interface{}) {
	wr.logger.Warningf("WAR %s %s:%d %s \n", v...)
}

func Error(v ...interface{}) {
	wr.logger.Errorf("ERR %s %s:%d %s \n", v...)
}

func Customize(color Color, prefix string, format string, v ...interface{}) {
	file, line := caller.Caller(1)

	var t = time.Now()

	color.Printf(format, v...)

	if DefaultLogger.Hook != nil {
		DefaultLogger.Hook(prefix, t, file, line, v...)
	}
}

func Assert(v ...interface{}) {
	if len(v) == 0 {
		return
	}
	if exception.IsNil(v[len(v)-1]) {
		return
	}

	wr.logger.Errorf("ERR %s %s:%d %s \n", v...)
}
