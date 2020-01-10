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
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type fi int

const File fi = iota

type Info struct {
	bytes []byte
	err   error
}

func (fi fi) ReadFromBytes(bts []byte) Info {
	return Info{err: nil, bytes: bts}
}

func (fi fi) ReadFromReader(r io.Reader) Info {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return Info{err: err, bytes: nil}
	}
	return Info{err: nil, bytes: b}
}

func (fi fi) ReadFromPath(path string) Info {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return Info{err: err, bytes: nil}
	}
	b, err := ioutil.ReadFile(absPath)
	if err != nil {
		return Info{err: err, bytes: nil}
	}
	return Info{err: nil, bytes: b}
}

func (i Info) LastError() error {
	return i.err
}

func (i Info) Bytes() []byte {
	return i.bytes
}

func (i Info) String() string {
	return string(i.bytes)
}

func (i Info) WriteToPath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	f, err := os.Create(absPath)
	defer func() { _ = f.Close() }()
	if err != nil {
		return err
	}

	_, err = io.Copy(f, bytes.NewReader(i.bytes))
	if err != nil {
		return err
	}

	return nil
}

func (i Info) WriteToReader(w io.Writer) error {
	_, err := io.Copy(w, bytes.NewReader(i.bytes))
	if err != nil {
		return err
	}
	return nil
}

func (i Info) WriteToBytes(bts []byte) error {
	if len(bts) <= len(i.bytes) {
		copy(bts, i.bytes)
		return nil
	}
	bts = bts[0 : len(i.bytes)-1]
	copy(bts, i.bytes)
	return nil
}
