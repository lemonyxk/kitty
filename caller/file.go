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
	"os"
	"runtime"
)

var absFilePath = false

func SetAbsFilePath(v bool) {
	absFilePath = v
}

func RuntimeCaller(deep int) (string, int) {
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
		file = file[len(rootPath)+1:]
	}
	return file, line
}
