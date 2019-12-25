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
	OnOpen    func(w http.ResponseWriter, r *http.Request)
	OnMessage func(stream *Stream)
	OnClose   func(w http.ResponseWriter, r *http.Request)
	OnError   func(err exception.ErrorFunc)

	middleware []func(stream *Stream) exception.ErrorFunc
	router     *HttpServerRouter
}

func (h *HttpServer) Ready() {

}

func (h *HttpServer) Use(middleware ...func(stream *Stream) exception.ErrorFunc) {
	h.middleware = append(h.middleware, middleware...)
}

func (h *HttpServer) handler(w http.ResponseWriter, r *http.Request) {

	if h.OnOpen != nil {
		h.OnOpen(w, r)
	}

	// Get the router
	node, formatPath := h.router.getRoute(r.Method, r.URL.Path)

	if node == nil {
		w.WriteHeader(http.StatusNotFound)
		if h.OnError != nil {
			h.OnError(exception.New("404 not found"))
		}
		if h.OnClose != nil {
			h.OnClose(w, r)
		}
		return
	}

	var params = &Params{Keys: node.Keys, Values: node.ParseParams(formatPath)}

	var stream = NewStream(w, r, params)

	var nodeData = node.Data.(*httpServerNode)

	if h.OnMessage != nil {
		h.OnMessage(stream)
	}

	for i := 0; i < len(h.middleware); i++ {
		err := h.middleware[i](stream)
		if err != nil {
			if h.OnError != nil {
				h.OnError(err)
			}
			if h.OnClose != nil {
				h.OnClose(w, r)
			}
			return
		}
	}

	for i := 0; i < len(nodeData.Before); i++ {
		context, err := nodeData.Before[i](stream)
		if err != nil {
			if h.OnError != nil {
				h.OnError(err)
			}
			if h.OnClose != nil {
				h.OnClose(w, r)
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
			if h.OnClose != nil {
				h.OnClose(w, r)
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
			if h.OnClose != nil {
				h.OnClose(w, r)
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
