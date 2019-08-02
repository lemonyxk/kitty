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

	// 挂载函数
	var f = http.HandlerFunc(sh)

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
			f := hh.GetRoute(strings.ToUpper(r.Method), r.URL.Path)
			if f == nil {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write(nil)
				return
			}

			// Get the middleware
			var context interface{}
			var err error
			var tool Stream

			tool.rs = rs{w, r, context}

			if hh.Middle != nil {
				context, err = hh.Middle(&tool)
				if err != nil {
					//log.Println(err)
					return
				}
				tool.Context = context
			}

			f(&tool)
		})
	}

	s.Run(handler(f))
}

// Start 启动
func (s *Server) Run(handler http.Handler) {

	if s.Protocol == "TLS" {
		log.Panicln(http.ListenAndServeTLS(fmt.Sprintf("%s:%d", s.Host, s.Port), s.CertFile, s.KeyFile, handler))
	}

	log.Panicln(http.ListenAndServe(fmt.Sprintf("%s:%d", s.Host, s.Port), handler))

}
