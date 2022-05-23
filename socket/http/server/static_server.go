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
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/lemonyxk/kitty/v2/errors"
	"github.com/lemonyxk/kitty/v2/kitty"
)

func (s *Server) staticHandler(w http.ResponseWriter, r *http.Request) error {

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

	return s.staticDefaultFileMiddle(w, err, file, ext)
}

func (s *Server) staticDefaultFileMiddle(w http.ResponseWriter, err error, file http.File, ext string) error {
	bts, err := ioutil.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return nil
	}

	var contentType = mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = kitty.TextPlain
	}

	w.Header().Set(kitty.ContentType, contentType)
	w.Header().Set(kitty.ContentLength, strconv.Itoa(len(bts)))
	_, err = w.Write(bts)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return nil
	}

	return nil
}

func (s *Server) staticDefaultDirMiddle(w http.ResponseWriter, r *http.Request, file http.File) error {
	dir, err := file.Readdir(0)
	if err != nil {
		return nil
	}

	var bts bytes.Buffer

	for i := 0; i < len(dir); i++ {
		var download = ""
		var empty = ""
		if s.staticRouter.staticDownload {
			download = `<a class="file" download href="` + filepath.Join(r.URL.Path, dir[i].Name()) + `">` + downloadSVG + `</a>`
			empty = emptySVG
		}
		if dir[i].IsDir() {
			bts.WriteString(`<div class="list">` + empty + dirSVG + `<a class="dir" href="` + filepath.Join(r.URL.Path, dir[i].Name()) + `">` + dir[i].Name() + `</a></div>`)
		} else {
			bts.WriteString(`<div class="list">` + download + fileSVG + `<a class="file" href="` + filepath.Join(r.URL.Path, dir[i].Name()) + `">` + dir[i].Name() + `</a>` + `</div>`)
		}
	}

	var str = strings.ReplaceAll(html, `{{body}}`, bts.String())

	w.Header().Set(kitty.ContentType, kitty.TextHtml)
	w.Header().Set(kitty.ContentLength, strconv.Itoa(len(str)))
	_, err = w.Write([]byte(str))
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return nil
	}

	return nil
}
