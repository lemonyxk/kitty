package lemo

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/Lemo-yxk/lemo/console"
)

// Server 服务结构
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

	server    *http.Server
	netListen net.Listener
}

type Handler struct {
	socketHandler *WebSocketServer
	httpHandler   *HttpServer
}

type Context interface{}

func (handler *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Match the websocket router
	if handler.socketHandler != nil && r.Method == http.MethodGet && handler.socketHandler.CheckPath(r.URL.Path, handler.socketHandler.Path) {
		handler.socketHandler.process(w, r)
		return
	}

	// HttpServer Not exists
	if handler.httpHandler == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// HttpServer router not exists
	if handler.httpHandler.router == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// static file
	if handler.httpHandler.router.staticPath != "" && r.Method == http.MethodGet {
		err := handler.httpHandler.staticHandler(w, r)
		if err == nil {
			return
		}
	}

	handler.httpHandler.handler(w, r)
	return
}

// Start 启动 WebSocketServer
func (s *Server) Start(socketHandler *WebSocketServer, httpHandler *HttpServer) {

	if socketHandler != nil {
		socketHandler.Ready()
	}

	if httpHandler != nil {
		httpHandler.Ready()
	}

	var handler = &Handler{
		socketHandler: socketHandler,
		httpHandler:   httpHandler,
	}

	s.run(handler)
}

// Start 启动
func (s *Server) run(handler http.Handler) {

	var server = http.Server{Addr: s.Host + ":" + strconv.Itoa(s.Port), Handler: handler}

	var err error
	var netListen net.Listener
	var isChild = os.Getenv("pid")

	if isChild != "" {
		f := os.NewFile(3, "")
		netListen, err = net.FileListener(f)
	} else {
		netListen, err = net.Listen("tcp", server.Addr)
	}

	if err != nil {
		panic(err)
	}

	s.netListen = netListen
	s.server = &server

	if s.Protocol == "TLS" {
		err = server.ServeTLS(netListen, s.CertFile, s.KeyFile)
	} else {
		err = server.Serve(netListen)
	}
	if err != nil {
		console.Error(err)
	}
}

func (s *Server) reload() {

	tl, ok := s.netListen.(*net.TCPListener)
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

func (s *Server) Shutdown() {
	err := s.server.Shutdown(context.Background())
	if err != nil {
		console.Error(err)
	}
	console.Println("kill pid:", os.Getpid())
}
