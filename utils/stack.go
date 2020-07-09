/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2020-07-09 20:49
**/

package utils

import (
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
)

var absFilePath = false

type stack int

const Stack stack = iota

func (s stack) SetAbsFilePath(v bool) {
	absFilePath = v
}

func (s stack) Caller(deep int) (string, int) {
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

func (s stack) Stack(deep int) (string, int) {
	var list = strings.Split(string(debug.Stack()), "\n")
	var info = strings.TrimSpace(list[deep])
	var flInfo = strings.Split(strings.Split(info, " ")[0], ":")
	var file, line = flInfo[0], flInfo[1]
	var l, _ = strconv.Atoi(line)
	return file, l
}

func (s stack) GetFuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}
