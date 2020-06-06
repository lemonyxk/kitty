package http

import (
	"context"
	"errors"
	"io/ioutil"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Lemo-yxk/lemo"
	"github.com/Lemo-yxk/lemo/console"
)

type Server struct {
	// Host 服务Host
	Host string
	// Port 服务端口
	Port int
	// Protocol 协议
	Protocol string
	// TLS FILE
	CertFile string
	// TLS KEY
	KeyFile string

	// AutoBind
	AutoBind bool

	OnOpen    func(stream *Stream)
	OnMessage func(stream *Stream)
	OnClose   func(stream *Stream)
	OnError   func(stream *Stream)

	middle []func(next Middle) Middle
	router *Router

	netListen net.Listener
	server    *http.Server
}

func (h *Server) Ready() {

}

type Middle func(*Stream)

func (h *Server) Use(middle ...func(next Middle) Middle) {
	h.middle = append(h.middle, middle...)
}

func (h *Server) process(w http.ResponseWriter, r *http.Request) {
	var stream = NewStream(h, w, r)
	h.middleware(stream)
}

func (h *Server) middleware(stream *Stream) {
	var next Middle = h.handler
	for i := len(h.middle) - 1; i >= 0; i-- {
		next = h.middle[i](next)
	}
	next(stream)
}

func (h *Server) handler(stream *Stream) {

	if h.OnOpen != nil {
		h.OnOpen(stream)
	}

	// Get the router
	n, formatPath := h.router.getRoute(stream.Request.Method, stream.Request.URL.Path)

	if n == nil {
		stream.Response.WriteHeader(http.StatusNotFound)
		stream.error = stream.Request.URL.Path + " " + "404 not found"
		if h.OnError != nil {
			h.OnError(stream)
		}
		if h.OnClose != nil {
			h.OnClose(stream)
		}
		return
	}

	stream.Params = lemo.Params{Keys: n.Keys, Values: n.ParseParams(formatPath)}

	var nodeData = n.Data.(*node)

	if h.OnMessage != nil {
		h.OnMessage(stream)
	}

	for i := 0; i < len(nodeData.before); i++ {
		ctx, err := nodeData.before[i](stream)
		if err != nil {
			stream.error = err
			if h.OnError != nil {
				h.OnError(stream)
			}
			if h.OnClose != nil {
				h.OnClose(stream)
			}
			return
		}
		stream.Context = ctx
	}

	if nodeData.function != nil {
		err := nodeData.function(stream)
		if err != nil {
			stream.error = err
			if h.OnError != nil {
				h.OnError(stream)
			}
			if h.OnClose != nil {
				h.OnClose(stream)
			}
			return
		}
	}

	for i := 0; i < len(nodeData.after); i++ {
		err := nodeData.after[i](stream)
		if err != nil {
			stream.error = err
			if h.OnError != nil {
				h.OnError(stream)
			}
			if h.OnClose != nil {
				h.OnClose(stream)
			}
			return
		}
	}
}

func (h *Server) staticHandler(w http.ResponseWriter, r *http.Request) error {

	if !strings.HasPrefix(r.URL.Path, h.router.prefixPath) {
		return errors.New("not match")
	}

	var absFilePath = filepath.Join(h.router.staticPath, r.URL.Path[len(h.router.prefixPath):])

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

	f, err := os.Open(absFilePath)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return nil
	}

	bts, err := ioutil.ReadAll(f)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return nil
	}

	w.Header().Set("Content-Type", contentType)
	_, err = w.Write(bts)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return nil
	}

	return nil

}

func (h *Server) SetRouter(router *Router) *Server {
	h.router = router
	return h
}

func (h *Server) GetRouter() *Router {
	return h.router
}

// Start Http
func (h *Server) Start() {

	h.Ready()

	var server = http.Server{Addr: h.Host + ":" + strconv.Itoa(h.Port), Handler: h}

	var err error
	var netListen net.Listener

	netListen, err = net.Listen("tcp", server.Addr)

	if err != nil {
		if strings.HasSuffix(err.Error(), "address already in use") {
			if h.AutoBind {
				h.Port++
				h.Start()
				return
			}
		}
		panic(err)
	}

	h.netListen = netListen
	h.server = &server

	switch h.Protocol {
	case "TLS":
		err = server.ServeTLS(netListen, h.CertFile, h.KeyFile)
	default:
		err = server.Serve(netListen)
	}

	console.Exit(err)
}

func (h *Server) Shutdown() {
	err := h.server.Shutdown(context.Background())
	if err != nil {
		panic(err)
	}
}

func (h *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// router not exists
	if h.router == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// static file
	if h.router.staticPath != "" && r.Method == http.MethodGet {
		err := h.staticHandler(w, r)
		if err == nil {
			return
		}
	}

	h.process(w, r)
	return
}
