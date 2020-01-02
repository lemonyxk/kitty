/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2020-01-02 16:08
**/

package utils

import (
	"os"
	"path/filepath"
)

type di int

const Dir di = iota

type dir struct {
	path string
	err  error
}

func (d di) New(path string) dir {
	var absPath, err = filepath.Abs(path)
	return dir{path: absPath, err: err}
}

func (d dir) LastError() error {
	return d.err
}

func (d dir) Walk() chan os.FileInfo {
	var ch = make(chan os.FileInfo)
	go func() {
		_ = filepath.Walk(d.path, func(path string, info os.FileInfo, err error) error {
			ch <- info
			return err
		})
		close(ch)
	}()
	return ch
}
