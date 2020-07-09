package console

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Lemo-yxk/lemo/exception"
	"github.com/Lemo-yxk/lemo/utils"
)

const (
	NONE   = 0
	STATUS = 1 << iota
	DATE
	FILE
)

var flag int

type Logger interface {
	Errorf(string, ...interface{})
	Warningf(string, ...interface{})
	Infof(string, ...interface{})
	Debugf(string, ...interface{})
}

type Formatter interface {
	Sprint(v ...interface{}) string
	Sprintf(format string, v ...interface{}) string
}

var wr *writer

type writer struct {
	logger    Logger
	formatter Formatter
}

func SetLogger(logger Logger) {
	wr.logger = logger
}

func SetFormatter(formatter Formatter) {
	wr.formatter = formatter
}

func SetFlags(v int) {
	flag = v
}

func init() {
	wr = &writer{}
	DefaultLogger = NewDefaultLogger()
	DefaultFormatter = NewDefaultFormatter()
	SetLogger(DefaultLogger)
	SetFormatter(DefaultFormatter)
	SetFlags(STATUS | DATE | FILE)
}

var DefaultLogger *defaultLogger

func NewDefaultLogger() *defaultLogger {
	return &defaultLogger{}
}

func NewDefaultFormatter() *defaultFormatter {
	return &defaultFormatter{}
}

var DefaultFormatter *defaultFormatter

type defaultFormatter struct{}

func (f *defaultFormatter) Sprintf(format string, v ...interface{}) string {
	return fmt.Sprintf(format, v...)
}

func (f *defaultFormatter) Sprint(v ...interface{}) string {
	return fmt.Sprint(v...)
}

type defaultLogger struct {
	Hook func(status string, t time.Time, file string, line int, v ...interface{})
}

func (log *defaultLogger) Errorf(format string, args ...interface{}) {

	var status = "ERR"

	var file string
	var line int
	var t time.Time
	var err interface{}
	if len(args) > 0 {
		err = args[0]
	}

	switch err.(type) {
	case exception.Error:
		file, line = err.(exception.Error).File(), err.(exception.Error).Line()
		t = err.(exception.Error).Time()
	default:
		file, line = utils.Stack.Caller(2)
		t = time.Now()
	}

	var flags []string

	if flag&STATUS != 0 {
		flags = append(flags, status)
	}

	if flag&DATE != 0 {
		flags = append(flags, t.Format("2006-01-02 15:04:05"))
	}

	if flag&FILE != 0 {
		flags = append(flags, file+":"+strconv.Itoa(line))
	}

	if len(flags) > 0 {
		format = "%s " + format
		args = append([]interface{}{strings.Join(flags, " ")}, args...)
	}

	FgRed.Printf(format, args...)

	if log.Hook != nil {
		log.Hook(status, t, file, line, args...)
	}

}

func (log *defaultLogger) Warningf(format string, args ...interface{}) {

	var status = "WAR"

	var t = time.Now()

	var flags []string

	var file, line = utils.Stack.Caller(2)

	if flag&STATUS != 0 {
		flags = append(flags, status)
	}

	if flag&DATE != 0 {
		flags = append(flags, t.Format("2006-01-02 15:04:05"))
	}

	if flag&FILE != 0 {
		flags = append(flags, file+":"+strconv.Itoa(line))
	}

	if len(flags) > 0 {
		format = "%s " + format
		args = append([]interface{}{strings.Join(flags, " ")}, args...)
	}

	FgYellow.Printf(format, args...)

	if log.Hook != nil {
		log.Hook(status, t, file, line, args...)
	}
}

func (log *defaultLogger) Infof(format string, args ...interface{}) {

	var status = "INF"

	var t = time.Now()

	var flags []string

	var file, line = utils.Stack.Caller(2)

	if flag&STATUS != 0 {
		flags = append(flags, status)
	}

	if flag&DATE != 0 {
		flags = append(flags, t.Format("2006-01-02 15:04:05"))
	}

	if flag&FILE != 0 {
		flags = append(flags, file+":"+strconv.Itoa(line))
	}

	if len(flags) > 0 {
		format = "%s " + format
		args = append([]interface{}{strings.Join(flags, " ")}, args...)
	}

	Bold.Printf(format, args...)

	if log.Hook != nil {
		log.Hook(status, t, file, line, args...)
	}
}

func (log *defaultLogger) Debugf(format string, args ...interface{}) {

	var status = "DEB"

	var t = time.Now()

	var flags []string

	var file, line = utils.Stack.Caller(2)

	if flag&STATUS != 0 {
		flags = append(flags, status)
	}

	if flag&DATE != 0 {
		flags = append(flags, t.Format("2006-01-02 15:04:05"))
	}

	if flag&FILE != 0 {
		flags = append(flags, file+":"+strconv.Itoa(line))
	}

	if len(flags) > 0 {
		format = "%s " + format
		args = append([]interface{}{strings.Join(flags, " ")}, args...)
	}

	FgBlue.Printf(format, args...)

	if log.Hook != nil {
		log.Hook(status, t, file, line, args...)
	}
}

func Exit(v interface{}) {
	wr.logger.Errorf("%v\n", v)
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

func Info(v ...interface{}) {
	var msg = utils.String.JoinInterface(v, " ")
	wr.logger.Infof("%s\n", msg)
}

func Debug(v ...interface{}) {
	var msg = utils.String.JoinInterface(v, " ")
	wr.logger.Debugf("%s\n", msg)
}

func Warning(v ...interface{}) {
	var msg = utils.String.JoinInterface(v, " ")
	wr.logger.Warningf("%s\n", msg)
}

func Error(v ...interface{}) {
	var err interface{}
	if len(v) > 0 {
		err = v[len(v)-1]
	}
	wr.logger.Errorf("%v\n", err)
}

func Infof(format string, v ...interface{}) {
	wr.logger.Infof(format, v...)
}

func Warningf(format string, v ...interface{}) {
	wr.logger.Warningf(format, v...)
}

func Debugf(format string, v ...interface{}) {
	wr.logger.Debugf(format, v...)
}

func Errorf(format string, v ...interface{}) {
	wr.logger.Errorf(format, v...)
}

func Assert(v ...interface{}) {
	if len(v) == 0 {
		return
	}
	if exception.IsNil(v[len(v)-1]) {
		return
	}

	wr.logger.Errorf("%v\n", v[len(v)-1])
}
