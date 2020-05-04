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
	"bytes"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
)

type zip int

const Zip zip = iota

type unzipFile struct {
	src string
	err error
}

type unzipReader struct {
	src io.ReaderAt
	err error
	len int64
}

func (z zip) UnzipFromFile(src string) unzipFile {
	var absPath, err = filepath.Abs(src)
	return unzipFile{src: absPath, err: err}
}

func (z zip) UnzipFromReader(src io.ReaderAt, len int64) unzipReader {
	return unzipReader{src: src, len: len, err: nil}
}

func (z zip) UnzipFromBytes(src []byte) unzipReader {
	var reader = bytes.NewReader(src)
	return unzipReader{src: reader, len: reader.Size(), err: nil}
}

func (z unzipReader) To(dst string) error {

	var absPath, err = filepath.Abs(dst)
	if err != nil {
		return err
	}

	_ = os.MkdirAll(absPath, 0666)

	cf, err := zip2.NewReader(z.src, z.len)
	if err != nil {
		return err
	}

	return doUnzip(&zip2.ReadCloser{
		Reader: *cf,
	}, absPath)
}

// need a dir
func (z unzipFile) To(dst string) error {

	var absPath, err = filepath.Abs(dst)
	if err != nil {
		return err
	}

	_ = os.MkdirAll(absPath, 0666)

	cf, err := zip2.OpenReader(z.src) // 读取zip文件
	if err != nil {
		return err
	}
	defer func() { _ = cf.Close() }()

	return doUnzip(cf, absPath)
}

func doUnzip(cf *zip2.ReadCloser, absPath string) error {
	for _, file := range cf.File {

		if file.FileInfo().IsDir() {
			_ = os.Mkdir(path.Join(absPath, file.Name), 0666)
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

type zipDir struct {
	src string
	err error
}

type zipFile struct {
	src string
	err error
}

// to zip
func (z zip) ZipFromDir(src string) zipDir {
	var absPath, err = filepath.Abs(src)
	return zipDir{src: absPath, err: err}
}

func (z zip) ZipFromFile(src string) zipFile {
	var absPath, err = filepath.Abs(src)
	return zipFile{src: absPath, err: err}
}

// need a file
func (z zipDir) To(dst string) error {

	var absPath, err = filepath.Abs(dst)
	if err != nil {
		return err
	}

	_ = os.MkdirAll(filepath.Dir(absPath), 0666)

	if _, err := os.Stat(filepath.Join(z.src, filepath.Base(absPath))); err == nil {
		return errors.New(absPath + " is exists")
	}

	fStat, err := os.Stat(z.src)
	if err != nil {
		return err
	}

	if !fStat.IsDir() {
		return errors.New(z.src + " is not dir")
	}

	var files = Dir.New(z.src).ReadAll()

	fw, err := os.Create(absPath)
	defer func() { _ = fw.Close() }()
	if err != nil {
		return err
	}

	var zw = zip2.NewWriter(fw)
	defer func() { _ = zw.Close() }()

	for i := 0; i < len(files); i++ {
		err = files[i].LastError()
		if err != nil {
			return err
		}
		err = doZip(files[i].Path(), files[i].Info(), zw)
		if err != nil {
			return err
		}
	}
	return err

}

func (z zipFile) To(dst string) error {

	var absPath, err = filepath.Abs(dst)
	if err != nil {
		return err
	}

	_ = os.MkdirAll(filepath.Dir(absPath), 0666)

	fStat, err := os.Stat(z.src)
	if err != nil {
		return err
	}

	if fStat.IsDir() {
		return errors.New(z.src + " is not file")
	}

	fw, err := os.Create(absPath)
	defer func() { _ = fw.Close() }()
	if err != nil {
		return err
	}

	var zw = zip2.NewWriter(fw)
	defer func() { _ = zw.Close() }()

	return doZip(z.src, fStat, zw)
}

func doZip(path string, fi os.FileInfo, zw *zip2.Writer) error {
	fh, err := zip2.FileInfoHeader(fi)
	if err != nil {
		return err
	}

	fh.Name = filepath.Base(path) + "2"

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
}
