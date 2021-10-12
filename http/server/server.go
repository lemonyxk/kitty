package server

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	http2 "github.com/lemoyxk/kitty/http"
	"github.com/lemoyxk/kitty/kitty"
)

type Server struct {
	Name string
	// Host 服务Host
	Addr string
	// Protocol 协议
	TLS bool
	// TLS FILE
	CertFile string
	// TLS KEY
	KeyFile string

	OnOpen    func(stream *http2.Stream)
	OnMessage func(stream *http2.Stream)
	OnClose   func(stream *http2.Stream)
	OnError   func(stream *http2.Stream, err error)
	OnSuccess func()

	middle    []func(next Middle) Middle
	router    *Router
	netListen net.Listener
	server    *http.Server
}

func (s *Server) Ready() {
	if s.Addr == "" {
		panic("Addr must set")
	}
}

type Middle func(*http2.Stream)

func (s *Server) LocalAddr() net.Addr {
	return s.netListen.Addr()
}

func (s *Server) Use(middle ...func(next Middle) Middle) {
	s.middle = append(s.middle, middle...)
}

func (s *Server) process(w http.ResponseWriter, r *http.Request) {
	var stream = &http2.Stream{Response: w, Request: r, Query: &http2.Store{}, Form: &http2.Store{}, Json: &http2.Json{}, Files: &http2.Files{}}
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
	n, formatPath := s.router.getRoute(stream.Request.Method, stream.Request.URL.Path)

	if n == nil {
		stream.Response.WriteHeader(http.StatusNotFound)
		var err = errors.New(stream.Request.URL.Path + " " + "404 not found")
		if s.OnError != nil {
			s.OnError(stream, err)
		}
		if s.OnClose != nil {
			s.OnClose(stream)
		}
		return
	}

	stream.Params = kitty.Params{Keys: n.Keys, Values: n.ParseParams(formatPath)}

	var nodeData = n.Data.(*node)

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

func (s *Server) staticHandler(w http.ResponseWriter, r *http.Request) error {

	if !strings.HasPrefix(r.URL.Path, s.router.prefixPath) {
		return errors.New("not match")
	}

	var absFilePath = filepath.Join(s.router.staticPath, r.URL.Path[len(s.router.prefixPath):])

	var info, err = os.Stat(absFilePath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		absFilePath = filepath.Join(absFilePath, s.router.defaultIndex)
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

func (s *Server) SetRouter(router *Router) *Server {
	s.router = router
	return s
}

func (s *Server) GetRouter() *Router {
	return s.router
}

// Start Http
func (s *Server) Start() {

	s.Ready()

	var server = http.Server{Addr: s.Addr, Handler: s}

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

	if s.TLS {
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

	// router not exists
	if s.router == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// static file
	if s.router.staticPath != "" && r.Method == http.MethodGet {
		err := s.staticHandler(w, r)
		if err == nil {
			return
		}
	}

	s.process(w, r)
	return
}
