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

type Multipart struct {
	Form  *Store
	Files *Files
}

type Files struct {
	files map[string][]*multipart.FileHeader
}

func (f *Files) Has(fileName string) bool {
	for name := range f.files {
		if fileName == name {
			return true
		}
	}

	return false
}

func (f *Files) Empty(fileName string) bool {
	var v = f.First(fileName)
	return v == nil
}

func (f *Files) Files() map[string][]*multipart.FileHeader {
	return f.files
}

func (f *Files) First(fileName string) *multipart.FileHeader {
	if file, ok := f.files[fileName]; ok {
		return file[0]
	}
	return nil
}

func (f *Files) Index(fileName string, index int) *multipart.FileHeader {
	if file, ok := f.files[fileName]; ok {
		return file[index]
	}
	return nil
}

func (f *Files) All(fileName string) []*multipart.FileHeader {
	if file, ok := f.files[fileName]; ok {
		return file
	}
	return nil
}

func (f *Files) String() string {

	var buff bytes.Buffer

	for name := range f.files {
		buff.WriteString(name + ":")
		for j := 0; j < len(f.files[name]); j++ {
			buff.WriteString(f.files[name][j].Filename)
			if j != len(f.files[name])-1 {
				buff.WriteString(",")
			}
		}
		buff.WriteString(" ")
	}

	if buff.Len() == 0 {
		return ""
	}

	var res = buff.String()

	return res[:len(res)-1]
}
