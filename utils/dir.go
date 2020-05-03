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

func (d dir) RemoveAll() error {
	return os.RemoveAll(d.path)
}

func (d dir) CreateAll(perm os.FileMode) error {
	return os.MkdirAll(d.path, perm)
}

func (d dir) Create(perm os.FileMode) error {
	return os.Mkdir(d.path, perm)
}

func (d dir) Exists() bool {
	_, err := os.Stat(d.path)
	return err == nil
}

func (d dir) LastError() error {
	return d.err
}

func (d dir) Walk() chan fileInfo {
	var ch = make(chan fileInfo)
	go func() {
		_ = filepath.Walk(d.path, func(path string, info os.FileInfo, err error) error {
			ch <- fileInfo{path, info, err}
			return err
		})
		close(ch)
	}()
	return ch
}

type fileInfo struct {
	path string
	info os.FileInfo
	err  error
}

func (f *fileInfo) LastError() error {
	return f.err
}

func (f *fileInfo) Info() os.FileInfo {
	return f.info
}

func (f *fileInfo) Path() string {
	return f.path
}
