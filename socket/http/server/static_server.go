/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2022-03-03 23:27
**/

package server

import (
	"bytes"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/lemonyxk/kitty/errors"
	"github.com/lemonyxk/kitty/kitty/header"
)

func (s *Server[T]) staticHandler(w http.ResponseWriter, r *http.Request) error {

	var static *Static

	var file http.File

	var err error

	var openPath string

	var urlPath string

	for i := 0; i < len(s.staticRouter.static); i++ {
		if !strings.HasPrefix(r.URL.Path, s.staticRouter.static[i].prefixPath) {
			continue
		}

		urlPath = r.URL.Path[len(s.staticRouter.static[i].prefixPath):]

		openPath = filepath.Join(s.staticRouter.static[i].fixPath, urlPath)

		file, err = s.staticRouter.static[i].fileSystem.Open(openPath)
		if err != nil {
			continue
		}

		static = s.staticRouter.static[i]

		break
	}

	if static == nil {
		return errors.Wrap(errors.NilError, "static")
	}

	if static.fileSystem == nil {
		return errors.Wrap(errors.NilError, "static file system")
	}

	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil {
		return nil
	}

	if info.IsDir() {

		var findDefault = false

		for i := 0; i < len(s.staticRouter.defaultIndex); i++ {
			if s.staticRouter.defaultIndex[i] == "" {
				continue
			}

			var otp = filepath.Join(openPath, s.staticRouter.defaultIndex[i])
			var of, err = static.fileSystem.Open(otp)
			if err != nil {
				continue
			}

			in, err := of.Stat()
			if err != nil {
				_ = of.Close()
				continue
			} else {
				_ = file.Close()
				openPath = otp
				file = of
				info = in
				findDefault = true
				break
			}
		}

		if !findDefault {

			if len(s.staticRouter.openDir) == 0 {
				w.WriteHeader(http.StatusForbidden)
				return nil
			}

			var shouldOpen = false
			for i := 0; i < len(s.staticRouter.openDir); i++ {
				if s.staticRouter.openDir[i] == static.index {
					shouldOpen = true
					break
				}
			}

			if !shouldOpen {
				w.WriteHeader(http.StatusForbidden)
				return nil
			}

			var fn, ok = s.staticRouter.staticDirMiddle[urlPath]
			if ok {
				var err = fn(w, r, file, info)
				if err != nil {
					w.WriteHeader(http.StatusForbidden)
					return nil
				}
				return nil
			}

			if s.staticRouter.staticGlobalDirMiddle != nil {
				var err = s.staticRouter.staticGlobalDirMiddle(w, r, file, info)
				if err != nil {
					w.WriteHeader(http.StatusForbidden)
					return nil
				}
				return nil
			}

			return s.staticDefaultDirMiddle(w, r, file)
		}
	}

	var ext = filepath.Ext(info.Name())

	var fn, ok = s.staticRouter.staticFileMiddle[ext]
	if ok {
		var err = fn(w, r, file, info)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return nil
		}
		return nil
	}

	if s.staticRouter.staticGlobalFileMiddle != nil {
		var err = s.staticRouter.staticGlobalFileMiddle(w, r, file, info)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return nil
		}
		return nil
	}

	return s.staticDefaultFileMiddle(w, err, file, info, ext)
}

func (s *Server[T]) staticDefaultFileMiddle(w http.ResponseWriter, err error, file http.File, info fs.FileInfo, ext string) error {
	var contentType = mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = header.TextPlain
	}

	w.Header().Set(header.ContentType, contentType)
	w.Header().Set(header.ContentLength, strconv.Itoa(int(info.Size())))
	_, err = io.Copy(w, file)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return nil
	}

	return nil
}

func (s *Server[T]) staticDefaultDirMiddle(w http.ResponseWriter, r *http.Request, file http.File) error {
	dir, err := file.Readdir(-1)
	if err != nil {
		return nil
	}

	var bts bytes.Buffer

	// title
	bts.WriteString(`<title>Index of ` + r.URL.Path + `</title>`)
	bts.WriteString(`<h1>Index of ` + r.URL.Path + `</h1>`)
	bts.WriteString(`<hr/>`)
	// pre
	bts.WriteString(`<pre>`)
	// back
	bts.WriteString(`<a href="../">` + "../" + `</a>` + "\n")

	for i := 0; i < len(dir); i++ {
		var p = filepath.Join(r.URL.Path, dir[i].Name())
		var name = dir[i].Name()
		var size = strconv.Itoa(int(dir[i].Size()))
		if dir[i].IsDir() {
			name = name + "/"
			p = p + "/"
			size = "-"
		}

		if len(name) > 50 {
			name = name[:47] + "..>"
		}
		var l = 50 - len(name)
		bts.WriteString(`<a href="` + p + `">` + name + `</a>` + strings.Repeat(" ", l))
		bts.WriteString(" " + dir[i].ModTime().Format("02-Jan-2006 15:04") + strings.Repeat(" ", 20-len(size)) + size)

		if s.staticRouter.staticDownload && !dir[i].IsDir() {
			bts.WriteString("  " + `<a download href="` + filepath.Join(r.URL.Path, dir[i].Name()) + `">` + "download" + `</a>`)
		}

		bts.WriteString("\n")
	}

	bts.WriteString(`</pre>`)
	bts.WriteString(`<hr/>`)

	w.Header().Set(header.ContentType, header.TextHtml)
	w.Header().Set(header.ContentLength, strconv.Itoa(len(bts.String())))
	_, err = w.Write([]byte(bts.String()))
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return nil
	}

	return nil
}
