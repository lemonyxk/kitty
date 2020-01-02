/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2020-01-02 15:59
**/

package utils

import (
	"io/ioutil"
	"path/filepath"
)

type fi int

const File fi = iota

type file struct {
	bytes []byte
	err   error
	path  string
	dir   string
	base  string
}

func (f fi) New(path string) file {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return file{err: err, bytes: nil}
	}
	bytes, err := ioutil.ReadFile(absPath)
	if err != nil {
		return file{err: err, bytes: nil}
	}
	return file{err: nil, bytes: bytes, dir: filepath.Dir(absPath), path: absPath, base: filepath.Base(absPath)}
}

func (f file) LastError() error {
	return f.err
}

func (f file) Bytes() []byte {
	return f.bytes
}

func (f file) String() string {
	return string(f.bytes)
}
