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

import "mime/multipart"

type Files struct {
	files map[string][]*multipart.FileHeader
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
