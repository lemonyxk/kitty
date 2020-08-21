/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2020-07-19 18:11
**/

package kitty

import (
	"os"
	"runtime"
	"strings"
)

func Caller(deep int) (string, int) {
	_, file, line, ok := runtime.Caller(deep + 1)
	if !ok {
		return "", 0
	}

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

	return file, line
}
