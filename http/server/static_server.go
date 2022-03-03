/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2022-03-03 23:27
**/

package server

import (
	"bytes"
	"errors"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/lemoyxk/kitty/kitty"
)

func (s *Server) staticHandler(w http.ResponseWriter, r *http.Request) error {

	if !strings.HasPrefix(r.URL.Path, s.router.prefixPath) {
		return errors.New("not match")
	}

	if s.router.fileSystem == nil {
		return errors.New("file system is nil")
	}

	var openPath = r.URL.Path[len(s.router.prefixPath):]

	openPath = filepath.Join(s.router.fixPath, openPath)

	var file, err = s.router.fileSystem.Open(openPath)
	if err != nil {
		return errors.New("not found")
	}
	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil {
		return nil
	}

	if info.IsDir() {

		var findDefault = false

		for i := 0; i < len(s.router.defaultIndex); i++ {
			if s.router.defaultIndex[i] == "" {
				continue
			}

			var otp = filepath.Join(openPath, s.router.defaultIndex[i])
			var of, err = s.router.fileSystem.Open(otp)
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

		if !findDefault && s.router.openDir {

			if s.router.dirMiddle != nil {
				var fn = s.router.dirMiddle
				err = fn(w, r, file, info)
				if err != nil {
					w.WriteHeader(http.StatusForbidden)
					return nil
				}
				return nil
			}

			dir, err := file.Readdir(0)
			if err != nil {
				return nil
			}

			var bts bytes.Buffer

			for i := 0; i < len(dir); i++ {
				var download = ""
				var empty = ""
				if s.router.staticDownload {
					download = `<a class="file" download href="` + filepath.Join(r.URL.Path, dir[i].Name()) + `">` + downloadSVG + `</a>`
					empty = emptySVG
				}
				if dir[i].IsDir() {
					bts.WriteString(`<div>` + empty + dirSVG + `<a class="dir" href="` + filepath.Join(r.URL.Path, dir[i].Name()) + `">` + dir[i].Name() + `</a></div>`)
				} else {
					bts.WriteString(`<div>` + download + fileSVG + `<a class="file" href="` + filepath.Join(r.URL.Path, dir[i].Name()) + `">` + dir[i].Name() + `</a>` + `</div>`)
				}
			}

			// bts.WriteString(`<div class="back"><a class="dir" href="` + filepath.Dir(r.URL.Path) + `">` + backSVG + `</a></div>`)

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
	}

	var ext = filepath.Ext(info.Name())

	var fn, ok = s.router.staticMiddle[ext]

	if ok {
		err = fn(w, r, file, info)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return nil
		}
		return nil
	}

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
