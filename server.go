package ws

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

// Server 服务结构
type Server struct {
	// Host 服务Host
	Host string
	// Port 服务端口
	Port int
	// Path
	Path string
	// Protocol 协议
	Protocol string
	// TLS FILE
	CertFile string
	// TLS KEY
	KeyFile string
}

func (s *Server) CatchError() {
	if err := recover(); err != nil {
		log.Println(string(debug.Stack()), err)
	}
}

// Start 启动 WebSocket
func (s *Server) Start(sh *Socket, hh *Http) {

	var ss = WebSocket(sh)

	// 中间件函数
	var handler = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			defer s.CatchError()

			var socketPath = r.URL.Path
			var httpPath = r.URL.Path
			var serverPath = s.Path

			// Match the websocket router
			if socketPath == serverPath {
				next.ServeHTTP(w, r)
				return
			}

			// Not exists
			if hh == nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write(nil)
				return
			}

			// Get the router
			hba, p := hh.GetRoute(r.Method, httpPath)
			if hba == nil {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write(nil)
				return
			}

			// Get the middleware
			var context interface{}
			var err error
			var tool Stream
			var params = p

			tool.rs = rs{w, r, context, params}

			for _, before := range hba.Before {
				context, err = before(&tool)
				if err != nil {
					return
				}
				tool.Context = context
			}

			hba.Handler(&tool)

			for _, after := range hba.After {
				_ = after(&tool)
			}
		})
	}

	s.Run(handler(ss))
}

// Start 启动
func (s *Server) Run(handler http.Handler) {
	if s.Protocol == "TLS" {
		log.Panicln(http.ListenAndServeTLS(fmt.Sprintf("%s:%d", s.Host, s.Port), s.CertFile, s.KeyFile, handler))
	} else {
		log.Panicln(http.ListenAndServe(fmt.Sprintf("%s:%d", s.Host, s.Port), handler))
	}
}
