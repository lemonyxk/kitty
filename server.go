package ws

import (
	"fmt"
	"log"
	"net/http"
	"strings"
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
		log.Println(err)
	}
}

// Start 启动 WebSocket
func (s *Server) Start(sh http.HandlerFunc, hh *HttpHandle) {

	// 中间件函数
	var handler = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			defer s.CatchError()

			// Match the websocket router
			if r.URL.Path == s.Path {
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
			hba := hh.GetRoute(strings.ToUpper(r.Method), r.URL.Path)
			if hba == nil {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write(nil)
				return
			}

			// Get the middleware
			var context interface{}
			var err error
			var tool Stream

			tool.rs = rs{w, r, context}

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

	s.Run(handler(sh))
}

// Start 启动
func (s *Server) Run(handler http.Handler) {
	if s.Protocol == "TLS" {
		log.Panicln(http.ListenAndServeTLS(fmt.Sprintf("%s:%d", s.Host, s.Port), s.CertFile, s.KeyFile, handler))
	} else {
		log.Panicln(http.ListenAndServe(fmt.Sprintf("%s:%d", s.Host, s.Port), handler))
	}
}
