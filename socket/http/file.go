/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2021-05-13 19:30
**/

package http

import (
	"bytes"
	"mime/multipart"
)

type FileHeader map[string][]*multipart.FileHeader

type File struct {
	FileHeader
}

func (f *File) Has(fileName string) bool {
	for name := range f.FileHeader {
		if fileName == name {
			return true
		}
	}

	return false
}

func (f *File) Empty(fileName string) bool {
	var v = f.First(fileName)
	return v == nil
}

func (f *File) Files() map[string][]*multipart.FileHeader {
	return f.FileHeader
}

func (f *File) First(fileName string) *multipart.FileHeader {
	if file, ok := f.FileHeader[fileName]; ok {
		return file[0]
	}
	return nil
}

func (f *File) Name(fileName string) Value {
	var res Value
	if file, ok := f.FileHeader[fileName]; ok {
		res.v = &file[0].Filename
		return res
	}
	return res
}

func (f *File) Index(fileName string, index int) *multipart.FileHeader {
	if file, ok := f.FileHeader[fileName]; ok {
		return file[index]
	}
	return nil
}

func (f *File) All(fileName string) []*multipart.FileHeader {
	if file, ok := f.FileHeader[fileName]; ok {
		return file
	}
	return nil
}

func (f *File) String() string {

	var buff bytes.Buffer

	for name := range f.FileHeader {
		buff.WriteString(name + ":")
		for j := 0; j < len(f.FileHeader[name]); j++ {
			buff.WriteString(f.FileHeader[name][j].Filename)
			if j != len(f.FileHeader[name])-1 {
				buff.WriteString(",")
			}
		}
		buff.WriteString(" ")
	}

	if buff.Len() == 0 {
		return ""
	}

	var bts = buff.Bytes()

	return string(bts[:len(bts)-1])
}
