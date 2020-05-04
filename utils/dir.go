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
	"io/ioutil"
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

func (d dir) ReadAll() []fileInfo {

	var res []fileInfo

	var fn func(path string, res *[]fileInfo)

	fn = func(path string, res *[]fileInfo) {

		files, err := ioutil.ReadDir(path)
		if err != nil {
			*res = append(*res, fileInfo{path, nil, err})
			return
		}

		for i := 0; i < len(files); i++ {
			var fullPath = filepath.Join(path, files[i].Name())
			if files[i].IsDir() {
				fn(fullPath, res)
			}
			*res = append(*res, fileInfo{fullPath, files[i], nil})
		}
	}

	fn(d.path, &res)

	return res

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
