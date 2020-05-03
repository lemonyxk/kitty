/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2020-05-04 00:33
**/

package utils

import (
	zip2 "archive/zip"
	"io"
	"os"
	"path"
	"path/filepath"
)

type zip int

const Zip zip = iota

type zipInfo struct {
	src string
	err error
}

func (z zip) New(src string) zipInfo {
	var absPath, err = filepath.Abs(src)
	return zipInfo{src: absPath, err: err}
}

func (z zipInfo) Unzip(dst string) error {

	var absPath, err = filepath.Abs(dst)
	if err != nil {
		return err
	}

	_ = os.MkdirAll(absPath, 0777)

	cf, err := zip2.OpenReader(z.src) // 读取zip文件
	if err != nil {
		return err
	}

	defer func() { _ = cf.Close() }()

	for _, file := range cf.File {

		if file.FileInfo().IsDir() {
			_ = os.Mkdir(path.Join(absPath, file.Name), 0777)
			continue
		}

		rc, err := file.Open()
		if err != nil {
			return err
		}

		f, err := os.Create(path.Join(absPath, file.Name))
		_, err = io.Copy(f, rc)
		if err != nil {
			_ = rc.Close()
			return err
		}

		_ = rc.Close()
		_ = f.Close()

	}
	return nil
}

func (z zipInfo) Zip(dst string) error {

	var absPath, err = filepath.Abs(dst)
	if err != nil {
		return err
	}

	_ = os.MkdirAll(filepath.Dir(absPath), 0777)

	fw, err := os.Create(absPath)
	defer func() { _ = fw.Close() }()
	if err != nil {
		return err
	}

	zw := zip2.NewWriter(fw)
	defer func() { _ = zw.Close() }()

	return filepath.Walk(z.src, func(path string, fi os.FileInfo, errBack error) error {
		if errBack != nil {
			return errBack
		}

		fh, err := zip2.FileInfoHeader(fi)
		if err != nil {
			return err
		}

		fh.Name = filepath.Base(path)

		if fi.IsDir() {
			fh.Name += string(filepath.Separator)
		}

		w, err := zw.CreateHeader(fh)
		if err != nil {
			return err
		}

		if !fh.Mode().IsRegular() {
			return nil
		}

		fr, err := os.Open(path)
		defer func() { _ = fr.Close() }()
		if err != nil {
			return err
		}

		_, err = io.Copy(w, fr)
		if err != nil {
			return err
		}

		return nil
	})
}
