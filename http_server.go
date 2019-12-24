package lemo

import (
	"errors"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lemo-yxk/lemo/exception"
)

type HttpServer struct {
	OnError    func(err exception.ErrorFunc)
	MiddleWare func(stream *Stream)
	router     *HttpServerRouter
}

func (h *HttpServer) Ready() {

}

func (h *HttpServer) handler(w http.ResponseWriter, r *http.Request) {

	// Get the router
	node, formatPath := h.router.getRoute(r.Method, r.URL.Path)

	var params = new(Params)
	if node != nil {
		params.Keys = node.Keys
		params.Values = node.ParseParams(formatPath)
	}

	var stream = NewStream(w, r, params)

	if h.MiddleWare != nil {
		h.MiddleWare(stream)
	}

	if node == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var nodeData = node.Data.(*httpServerNode)

	// Get the middleware
	for i := 0; i < len(nodeData.Before); i++ {
		context, err := nodeData.Before[i](stream)
		if err != nil {
			if h.OnError != nil {
				h.OnError(err)
			}
			return
		}
		stream.Context = context
	}

	if nodeData.HttpServerFunction != nil {
		err := nodeData.HttpServerFunction(stream)
		if err != nil {
			if h.OnError != nil {
				h.OnError(err)
			}
			return
		}
	}

	for i := 0; i < len(nodeData.After); i++ {
		err := nodeData.After[i](stream)
		if err != nil {
			if h.OnError != nil {
				h.OnError(err)
			}
			return
		}
	}
}

func (h *HttpServer) staticHandler(w http.ResponseWriter, r *http.Request) error {

	if !strings.HasPrefix(r.URL.Path, h.router.prefixPath) {
		return errors.New("not match")
	}

	var absFilePath = h.router.staticPath + r.URL.Path[len(h.router.prefixPath):]

	var info, err = os.Stat(absFilePath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		absFilePath = filepath.Join(absFilePath, h.router.defaultIndex)
		if _, err := os.Stat(absFilePath); err != nil {
			return errors.New("staticPath is not a file")
		}
	}

	// has found
	var contentType = mime.TypeByExtension(filepath.Ext(absFilePath))

	f, err := os.OpenFile(absFilePath, os.O_RDONLY, 0666)
	if err != nil {
		if h.OnError != nil {
			h.OnError(exception.New(err))
		}
		w.WriteHeader(http.StatusForbidden)
		return nil
	}

	bts, err := ioutil.ReadAll(f)
	if err != nil {
		if h.OnError != nil {
			h.OnError(exception.New(err))
		}
		w.WriteHeader(http.StatusForbidden)
		return nil
	}

	w.Header().Set("Content-Type", contentType)
	_, err = w.Write(bts)
	if err != nil {
		if h.OnError != nil {
			h.OnError(exception.New(err))
		}
		w.WriteHeader(http.StatusForbidden)
		return nil
	}

	return nil

}

func (h *HttpServer) Router(router *HttpServerRouter) *HttpServer {
	h.router = router
	return h
}
