package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	http2 "github.com/lemoyxk/kitty/http"
	"github.com/lemoyxk/kitty/kitty"
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
	var stream = &http2.Stream{Response: w, Request: r, Protobuf: &http2.Protobuf{}, Query: &http2.Store{}, Form: &http2.Store{}, Json: &http2.Json{}, Files: &http2.Files{}}
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

	if s.router.fileSystem == nil {
		return errors.New("file system is nil")
	}

	var openPath = r.URL.Path[len(s.router.prefixPath):]

	openPath = filepath.Join(s.router.fixPath, openPath)

	var file, err = s.router.fileSystem.Open(openPath)
	if err != nil {
		return errors.New("not found")
	}
	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil {
		return nil
	}

	if info.IsDir() {

		var findDefault = false

		for i := 0; i < len(s.router.defaultIndex); i++ {
			if s.router.defaultIndex[i] == "" {
				continue
			}

			var otp = filepath.Join(openPath, s.router.defaultIndex[i])
			var of, err = s.router.fileSystem.Open(otp)
			if err != nil {
				continue
			}

			in, err := of.Stat()
			if err != nil {
				_ = of.Close()
				continue
			} else {
				_ = file.Close()
				openPath = otp
				file = of
				info = in
				findDefault = true
				break
			}
		}

		if !findDefault && s.router.openDir {

			if s.router.dirMiddle != nil {
				var fn = s.router.dirMiddle
				err = fn(w, r, file, info)
				if err != nil {
					w.WriteHeader(http.StatusForbidden)
					return nil
				}
				return nil
			}

			dir, err := file.Readdir(0)
			if err != nil {
				return nil
			}

			var bts bytes.Buffer

			bts.WriteString(`
				<!DOCTYPE html>
				<html lang="en">
				<head>
				    <meta charset="UTF-8">
				    <meta http-equiv="X-UA-Compatible" content="IE=edge">
				    <meta name="viewport" content="width=device-width, initial-scale=1.0">
					<style>
						.dir {
						    color: #008CBA;
						    line-height: inherit;
						    text-decoration: none;
						}
						.file {
						    color: #000000;
						    line-height: inherit;
						    text-decoration: none;
						}
						a:hover {
						    color: #008C0A;
						}
						div {
						    margin-top: 4px;
						    display: flex;
						    justify-content: flex-start;
						    align-items: center;
						}
						svg {
							width:1rem;
							margin-right: 5px;
						}
					</style>
				    <title>kitty-server</title>
				</head>
				<body>
			`)

			for i := 0; i < len(dir); i++ {
				if dir[i].IsDir() {
					bts.WriteString(`<div><svg class="MuiSvgIcon-root MuiSvgIcon-fontSizeMedium MuiSvgIcon-root MuiSvgIcon-fontSizeLarge css-zjt8k" focusable="false" aria-hidden="true" viewBox="0 0 24 24" data-testid="DriveFileMoveIcon" tabindex="-1" title="DriveFileMove"><path d="M20 6h-8l-2-2H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2zm-6 12v-3h-4v-4h4V8l5 5-5 5z"></path></svg><a class="dir" href="` + filepath.Join(r.URL.Path, dir[i].Name()) + `">` + dir[i].Name() + `</a></div>`)
				} else {
					bts.WriteString(`<div><svg class="MuiSvgIcon-root MuiSvgIcon-fontSizeMedium MuiSvgIcon-root MuiSvgIcon-fontSizeLarge css-zjt8k" focusable="false" aria-hidden="true" viewBox="0 0 24 24" data-testid="ArticleIcon" tabindex="-1" title="Article"><path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-5 14H7v-2h7v2zm3-4H7v-2h10v2zm0-4H7V7h10v2z"></path></svg><a class="file" href="` + filepath.Join(r.URL.Path, dir[i].Name()) + `">` + dir[i].Name() + `</a></div>`)
				}
			}

			bts.WriteString(`<div><a class="dir" href="` + filepath.Dir(r.URL.Path) + `">` + `<svg class="MuiSvgIcon-root MuiSvgIcon-fontSizeMedium MuiSvgIcon-root MuiSvgIcon-fontSizeLarge css-zjt8k" focusable="false" aria-hidden="true" viewBox="0 0 24 24" data-testid="ArrowBackIcon" tabindex="-1" title="ArrowBack"><path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z"></path></svg>` + `</a></div>`)

			bts.WriteString(`				    
				</body>
				</html>
			`)

			w.Header().Set(kitty.ContentType, kitty.TextHtml)
			w.Header().Set(kitty.ContentLength, strconv.Itoa(bts.Len()))
			_, err = w.Write(bts.Bytes())
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				return nil
			}

			return nil
		}
	}

	var ext = filepath.Ext(info.Name())

	var fn, ok = s.router.staticMiddle[ext]

	if ok {
		err = fn(w, r, file, info)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return nil
		}
		return nil
	}

	bts, err := ioutil.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return nil
	}

	var contentType = mime.TypeByExtension(ext)

	w.Header().Set(kitty.ContentType, contentType)
	w.Header().Set(kitty.ContentLength, strconv.Itoa(len(bts)))
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

	// router not exists
	if s.router == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// static file
	if s.router.fileSystem != nil && r.Method == http.MethodGet {
		if s.staticHandler(w, r) == nil {
			return
		}
	}

	s.process(w, r)
	return
}
