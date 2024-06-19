package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/lemonyxk/kitty/errors"
	"github.com/lemonyxk/kitty/router"
	http2 "github.com/lemonyxk/kitty/socket/http"
)

type Server[T any] struct {
	Name string
	// Host 服务Host
	Addr string
	// TLS FILE
	CertFile string
	// TLS KEY
	KeyFile string
	// TLS
	TLSConfig *tls.Config

	OnOpen    func(stream *http2.Stream[Conn])
	OnMessage func(stream *http2.Stream[Conn])
	OnClose   func(stream *http2.Stream[Conn])
	OnError   func(stream *http2.Stream[Conn], err error)
	OnRaw     func(w http.ResponseWriter, r *http.Request)
	OnSuccess func()

	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	MaxHeaderBytes    int

	middle       []func(next Middle) Middle
	router       *router.Router[*http2.Stream[Conn], T]
	staticRouter *StaticRouter
	netListen    net.Listener
	server       *http.Server
}

type Middle router.Middle[*http2.Stream[Conn]]

func (s *Server[T]) Ready() {
	if s.Addr == "" {
		panic("addr can not be empty")
	}
}

func (s *Server[T]) LocalAddr() net.Addr {
	return s.netListen.Addr()
}

func (s *Server[T]) Use(middle ...func(next Middle) Middle) {
	s.middle = append(s.middle, middle...)
}

func (s *Server[T]) process(w http.ResponseWriter, r *http.Request) {
	var stream = http2.NewStream[Conn](&conn{}, w, r)
	s.middleware(stream)
}

func (s *Server[T]) middleware(stream *http2.Stream[Conn]) {
	var next Middle = s.handler
	for i := len(s.middle) - 1; i >= 0; i-- {
		next = s.middle[i](next)
	}
	next(stream)
}

func (s *Server[T]) handler(stream *http2.Stream[Conn]) {

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

	stream.Params = n.ParseParams(formatPath)

	//stream.Node = n.Data

	var nodeData = n.Data

	if s.OnMessage != nil {
		s.OnMessage(stream)
	}

	for i := 0; i < len(nodeData.Before); i++ {
		if err := nodeData.Before[i](stream); err != nil {
			if errors.Is(err, errors.StopPropagation) {
				return
			}
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
			if errors.Is(err, errors.StopPropagation) {
				return
			}
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
			if errors.Is(err, errors.StopPropagation) {
				return
			}
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

func (s *Server[T]) SetRouter(router *router.Router[*http2.Stream[Conn], T]) *Server[T] {
	s.router = router
	return s
}

func (s *Server[T]) SetStaticRouter(router *StaticRouter) *Server[T] {
	s.staticRouter = router
	return s
}

func (s *Server[T]) GetRouter() *router.Router[*http2.Stream[Conn], T] {
	return s.router
}

// Start Http
func (s *Server[T]) Start() {

	s.Ready()

	var server = http.Server{
		Addr: s.Addr, Handler: s,
		ReadTimeout:       s.ReadTimeout,
		WriteTimeout:      s.WriteTimeout,
		IdleTimeout:       s.IdleTimeout,
		ReadHeaderTimeout: s.ReadHeaderTimeout,
		MaxHeaderBytes:    s.MaxHeaderBytes,
	}

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

	if s.KeyFile != "" && s.CertFile != "" || s.TLSConfig != nil {
		server.TLSConfig = s.TLSConfig
		err = server.ServeTLS(netListen, s.CertFile, s.KeyFile)
	} else {
		err = server.Serve(netListen)
	}

	if err != nil {
		fmt.Println(err)
	}
}

func (s *Server[T]) Shutdown() error {
	return s.server.Shutdown(context.Background())
}

func (s *Server[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet && r.Header.Get("Upgrade") == "websocket" {
		if s.OnRaw != nil {
			s.OnRaw(w, r)
			return
		}
	}

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
