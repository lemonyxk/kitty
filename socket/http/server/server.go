package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/lemonyxk/kitty/errors"
	"github.com/lemonyxk/kitty/router"
	"github.com/lemonyxk/kitty/socket"
	http2 "github.com/lemonyxk/kitty/socket/http"
)

type Server struct {
	Name string
	// Host 服务Host
	Addr string
	// TLS FILE
	CertFile string
	// TLS KEY
	KeyFile string

	OnOpen    func(stream *http2.Stream)
	OnMessage func(stream *http2.Stream)
	OnClose   func(stream *http2.Stream)
	OnError   func(stream *http2.Stream, err error)
	OnSuccess func()

	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	middle       []func(next Middle) Middle
	router       *router.Router[*http2.Stream]
	staticRouter *StaticRouter
	netListen    net.Listener
	server       *http.Server
}

type Middle router.Middle[*http2.Stream]

func (s *Server) Ready() {
	if s.Addr == "" {
		panic("addr can not be empty")
	}
}

func (s *Server) LocalAddr() net.Addr {
	return s.netListen.Addr()
}

func (s *Server) Use(middle ...func(next Middle) Middle) {
	s.middle = append(s.middle, middle...)
}

func (s *Server) process(w http.ResponseWriter, r *http.Request) {
	var stream = http2.NewStream(w, r)
	s.middleware(stream)
}

func (s *Server) middleware(stream *http2.Stream) {
	var next Middle = s.handler
	for i := len(s.middle) - 1; i >= 0; i-- {
		next = s.middle[i](next)
	}
	next(stream)
}

func (s *Server) handler(stream *http2.Stream) {

	if s.OnOpen != nil {
		s.OnOpen(stream)
	}

	// Get the router
	var method = strings.ToUpper(stream.Request.Method)
	n, formatPath := s.router.GetRoute(stream.Request.URL.Path)

	if n == nil {
		stream.Response.WriteHeader(http.StatusNotFound)
		var err = errors.Wrap(errors.RouteNotFount, stream.Request.URL.Path)
		if s.OnError != nil {
			s.OnError(stream, err)
		}
		if s.OnClose != nil {
			s.OnClose(stream)
		}
		return
	}

	var allowMethod = false
	for i := 0; i < len(n.Data.Method); i++ {
		if method == n.Data.Method[i] {
			allowMethod = true
			break
		}
	}

	if !allowMethod {
		stream.Response.WriteHeader(http.StatusMethodNotAllowed)
		var err = errors.Wrap(errors.MethodNotAllowed, stream.Request.URL.Path)
		if s.OnError != nil {
			s.OnError(stream, err)
		}
		if s.OnClose != nil {
			s.OnClose(stream)
		}
		return
	}

	stream.Params = socket.Params{Keys: n.Keys, Values: n.ParseParams(formatPath)}

	var nodeData = n.Data

	if s.OnMessage != nil {
		s.OnMessage(stream)
	}

	for i := 0; i < len(nodeData.Before); i++ {
		if err := nodeData.Before[i](stream); err != nil {
			if s.OnError != nil {
				s.OnError(stream, err)
			}
			if s.OnClose != nil {
				s.OnClose(stream)
			}
			return
		}
	}

	if nodeData.Function != nil {
		if err := nodeData.Function(stream); err != nil {
			if s.OnError != nil {
				s.OnError(stream, err)
			}
			if s.OnClose != nil {
				s.OnClose(stream)
			}
			return
		}
	}

	for i := 0; i < len(nodeData.After); i++ {
		if err := nodeData.After[i](stream); err != nil {
			if s.OnError != nil {
				s.OnError(stream, err)
			}
			if s.OnClose != nil {
				s.OnClose(stream)
			}
			return
		}
	}
}

func (s *Server) SetRouter(router *router.Router[*http2.Stream]) *Server {
	s.router = router
	return s
}

func (s *Server) SetStaticRouter(router *StaticRouter) *Server {
	s.staticRouter = router
	return s
}

func (s *Server) GetRouter() *router.Router[*http2.Stream] {
	return s.router
}

// Start Http
func (s *Server) Start() {

	s.Ready()

	var server = http.Server{Addr: s.Addr, Handler: s, ReadTimeout: s.ReadTimeout, WriteTimeout: s.WriteTimeout}

	var err error
	var netListen net.Listener

	netListen, err = net.Listen("tcp", server.Addr)

	if err != nil {
		panic(err)
	}

	s.netListen = netListen
	s.server = &server

	// start success
	if s.OnSuccess != nil {
		s.OnSuccess()
	}

	if s.KeyFile != "" && s.CertFile != "" {
		err = server.ServeTLS(netListen, s.CertFile, s.KeyFile)
	} else {
		err = server.Serve(netListen)
	}

	if err != nil {
		fmt.Println(err)
	}
}

func (s *Server) Shutdown() error {
	return s.server.Shutdown(context.Background())
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// static file
	if s.staticRouter != nil && s.staticRouter.IsAllowMethod(r.Method) {
		if s.staticHandler(w, r) == nil {
			return
		}
	}

	if s.router != nil {
		s.process(w, r)
		return
	}

	return
}
