/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-11-28 17:27
**/

package caller

import (
	"bytes"
	"os"
	"runtime"
	"strconv"
	"strings"
)

var absFilePath = false

func SetAbsFilePath(v bool) {
	absFilePath = v
}

type Trace struct {
	list []runtime.Frame
}

func (t *Trace) String() string {

	var buf bytes.Buffer

	for i := 0; i < len(t.list); i++ {
		buf.WriteString(strings.Repeat("  ", i))
		buf.WriteString(t.list[i].File + ":" + strconv.Itoa(t.list[i].Line))
		if i != len(t.list)-1 {
			buf.WriteString("\n")
		}
	}

	return buf.String()
}

func (t *Trace) All() []runtime.Frame {
	if !absFilePath {
		var res []runtime.Frame
		for i := 0; i < len(t.list); i++ {
			res = append(res, t.list[i])
			var rootPath, err = os.Getwd()
			if err != nil {
				continue
			}
			if rootPath == "/" {
				continue
			}
			if strings.HasPrefix(t.list[i].File, rootPath) {
				res[i].File = res[i].File[len(rootPath)+1:]
			}
		}
		return res
	}
	return t.list
}

func Stack() *Trace {
	var pcs = make([]uintptr, 32)
	n := runtime.Callers(1, pcs[:])

	frames := runtime.CallersFrames(pcs[:n])
	cs := make([]runtime.Frame, 0, n)

	frame, more := frames.Next()

	for more {
		frame, more = frames.Next()
		cs = append(cs, frame)
	}

	return &Trace{list: cs}
}

func Caller(deep int) (string, int) {
	_, file, line, ok := runtime.Caller(deep + 1)
	if !ok {
		return "", 0
	}
	if !absFilePath {
		var rootPath, err = os.Getwd()
		if err != nil {
			return file, line
		}
		if rootPath == "/" {
			return file, line
		}
		if strings.HasPrefix(file, rootPath) {
			file = file[len(rootPath)+1:]
		}
	}
	return file, line
}
