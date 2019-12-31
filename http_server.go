package lemo

import (
	"context"
	"errors"
	"io/ioutil"
	"mime"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Lemo-yxk/lemo/console"
)

type HttpServer struct {
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

	OnOpen    func(stream *Stream)
	OnMessage func(stream *Stream)
	OnClose   func(stream *Stream)
	OnError   func(stream *Stream)

	middle []func(next HttpServerMiddle) HttpServerMiddle
	router *HttpServerRouter

	netListen net.Listener
	server    *http.Server
}

func (h *HttpServer) Ready() {

}

type HttpServerMiddle func(*Stream)

func (h *HttpServer) Use(middle ...func(next HttpServerMiddle) HttpServerMiddle) {
	h.middle = append(h.middle, middle...)
}

func (h *HttpServer) process(w http.ResponseWriter, r *http.Request) {
	var stream = NewStream(h, w, r)
	h.middleware(stream)
}

func (h *HttpServer) middleware(stream *Stream) {
	var next HttpServerMiddle = h.handler
	for i := len(h.middle) - 1; i >= 0; i-- {
		next = h.middle[i](next)
	}
	next(stream)
}

func (h *HttpServer) handler(stream *Stream) {

	if h.OnOpen != nil {
		h.OnOpen(stream)
	}

	// Get the router
	node, formatPath := h.router.getRoute(stream.Request.Method, stream.Request.URL.Path)

	if node == nil {
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

	stream.Params = Params{Keys: node.Keys, Values: node.ParseParams(formatPath)}

	var nodeData = node.Data.(*httpServerNode)

	if h.OnMessage != nil {
		h.OnMessage(stream)
	}

	for i := 0; i < len(nodeData.Before); i++ {
		ctx, err := nodeData.Before[i](stream)
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

	if nodeData.HttpServerFunction != nil {
		err := nodeData.HttpServerFunction(stream)
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

	for i := 0; i < len(nodeData.After); i++ {
		err := nodeData.After[i](stream)
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

func (h *HttpServer) SetRouter(router *HttpServerRouter) *HttpServer {
	h.router = router
	return h
}

func (h *HttpServer) GetRouter() *HttpServerRouter {
	return h.router
}

// Start Http
func (h *HttpServer) Start() {

	h.Ready()

	var server = http.Server{Addr: h.Host + ":" + strconv.Itoa(h.Port), Handler: h}

	var err error
	var netListen net.Listener

	switch os.Getenv("pid") {
	case "":
		netListen, err = net.Listen("tcp", server.Addr)
	default:
		f := os.NewFile(3, "")
		netListen, err = net.FileListener(f)
	}

	if err != nil {
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

	if err != nil {
		console.Error(err)
	}
}

func (h *HttpServer) reload() {

	tl, ok := h.netListen.(*net.TCPListener)
	if !ok {
		panic("listener is not tcp listener")
	}

	f, err := tl.File()
	if err != nil {
		panic(err)
	}

	err = os.Setenv("pid", strconv.Itoa(os.Getpid()))
	if err != nil {
		panic(err)
	}

	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.ExtraFiles = []*os.File{f}
	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	console.Println("new pid:", cmd.Process.Pid)
}

func (h *HttpServer) Shutdown() {
	err := h.server.Shutdown(context.Background())
	if err != nil {
		console.Error(err)
	}
	console.Println("kill pid:", os.Getpid())
}

func (h *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// HttpServer router not exists
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
