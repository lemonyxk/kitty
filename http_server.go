package lemo

import (
	"errors"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lemo-yxk/tire"
)

type HttpServer struct {
	IgnoreCase bool
	Router     *tire.Tire
	OnError    ErrorFunction

	group        *httpServerGroup
	route        *httpServerRoute
	prefixPath   string
	staticPath   string
	defaultIndex string
}

func (h *HttpServer) Ready() {
	h.SetDefaultIndex("index.html")
}

func (h *HttpServer) SetDefaultIndex(index string) {
	h.defaultIndex = index
}

func (h *HttpServer) SetStaticPath(prefixPath string, staticPath string) {

	if prefixPath == "" {
		panic("prefixPath can not be empty")
	}

	if staticPath == "" {
		panic("staticPath can not be empty")
	}

	absStaticPath, err := filepath.Abs(staticPath)
	if err != nil {
		panic(err)
	}

	info, err := os.Stat(absStaticPath)
	if err != nil {
		panic(err)
	}

	if !info.IsDir() {
		panic("staticPath is not a dir")
	}

	h.prefixPath = prefixPath
	h.staticPath = absStaticPath
}

func (h *HttpServer) staticHandler(w http.ResponseWriter, r *http.Request) error {

	if !strings.HasPrefix(r.URL.Path, h.prefixPath) {
		return errors.New("not match")
	}

	var absFilePath = h.staticPath + r.URL.Path[len(h.prefixPath):]

	var info, err = os.Stat(absFilePath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		absFilePath = filepath.Join(absFilePath, h.defaultIndex)
		if _, err := os.Stat(absFilePath); err != nil {
			return errors.New("staticPath is not a file")
		}
	}

	// has found
	var contentType = mime.TypeByExtension(filepath.Ext(absFilePath))

	f, err := os.OpenFile(absFilePath, os.O_RDONLY, 0666)
	if err != nil {
		if h.OnError != nil {
			h.OnError(NewError(err))
		}
		w.WriteHeader(http.StatusForbidden)
		return nil
	}

	bts, err := ioutil.ReadAll(f)
	if err != nil {
		if h.OnError != nil {
			h.OnError(NewError(err))
		}
		w.WriteHeader(http.StatusForbidden)
		return nil
	}

	w.Header().Set("Content-Type", contentType)
	_, err = w.Write(bts)
	if err != nil {
		if h.OnError != nil {
			h.OnError(NewError(err))
		}
		w.WriteHeader(http.StatusForbidden)
		return nil
	}

	return nil

}
