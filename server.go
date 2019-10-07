package lemo

import (
	"fmt"
	"net/http"
)

// Version
var Version string = "1.0.1"

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
}

type Handler struct {
	socketHandler *Socket
	httpHandler   *Http
}

type Context interface{}

func (handler *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Match the websocket router
	if handler.socketHandler != nil && handler.socketHandler.CheckPath(r.URL.Path, handler.socketHandler.Path) {
		handler.socketHandler.upgrade(w, r)
		return
	}

	// Http Not exists
	if handler.httpHandler == nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(nil)
		return
	}

	handler.httpHandler.handle(w, r)
	return
}

// Start 启动 WebSocket
func (s *Server) Start(socketHandler *Socket, httpHandler *Http) {
	socketHandler.Init()

	var handler = &Handler{
		socketHandler: socketHandler,
		httpHandler:   httpHandler,
	}

	s.Run(handler)
}

// Start 启动
func (s *Server) Run(handler http.Handler) {

	switch s.Protocol {
	case "TLS":
		panic(http.ListenAndServeTLS(fmt.Sprintf("%s:%d", s.Host, s.Port), s.CertFile, s.KeyFile, handler))
	case "":
		panic(http.ListenAndServe(fmt.Sprintf("%s:%d", s.Host, s.Port), handler))
	}
}
