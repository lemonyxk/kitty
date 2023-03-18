package http

import (
	"bytes"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/lemonyxk/kitty/v2/kitty"
	"github.com/lemonyxk/kitty/v2/socket"
)

type Stream struct {
	// Server   *Server
	Response http.ResponseWriter
	Request  *http.Request

	Query     *Store
	Form      *Store
	Multipart *Multipart
	Json      *Json
	Protobuf  *Protobuf

	Params  socket.Params
	Context kitty.Context
	Logger  kitty.Logger

	Sender *Sender

	Parser *Parser
}

func NewStream(w http.ResponseWriter, r *http.Request) *Stream {
	var stream = &Stream{
		Response: w, Request: r,
		Protobuf:  &Protobuf{},
		Query:     &Store{},
		Form:      &Store{},
		Json:      &Json{},
		Multipart: &Multipart{Files: &Files{}, Form: &Store{}},
		Sender:    &Sender{response: w, request: r},
		Parser:    &Parser{response: w, request: r, maxMemory: 6 * 1024 * 1024},
	}

	stream.Parser.stream = stream

	return stream
}

func (s *Stream) Forward(fn func(stream *Stream) error) error {
	return fn(s)
}

func (s *Stream) SetHeader(header string, content string) {
	s.Response.Header().Set(header, content)
}

func (s *Stream) Host() string {
	if host := s.Request.Header.Get(kitty.Host); host != "" {
		return host
	}
	return s.Request.Host
}

func (s *Stream) ClientIP() string {

	if ip := strings.Split(s.Request.Header.Get(kitty.XForwardedFor), ",")[0]; ip != "" {
		return ip
	}

	if ip := s.Request.Header.Get(kitty.XRealIP); ip != "" {
		return ip
	}

	if ip, _, err := net.SplitHostPort(s.Request.RemoteAddr); err == nil {
		return ip
	}

	return ""
}

func (s *Stream) Has(key string) bool {
	if strings.ToUpper(s.Request.Method) == "GET" {
		return s.Query.Has(key)
	}

	var header = s.Request.Header.Get(kitty.ContentType)

	if strings.HasPrefix(header, kitty.MultipartFormData) {
		return s.Multipart.Form.Has(key)
	}

	if strings.HasPrefix(header, kitty.ApplicationFormUrlencoded) {
		return s.Form.Has(key)
	}

	if strings.HasPrefix(header, kitty.ApplicationJson) {
		return s.Json.Has(key)
	}

	return false
}

func (s *Stream) Empty(key string) bool {
	if strings.ToUpper(s.Request.Method) == "GET" {
		return s.Query.Empty(key)
	}

	var header = s.Request.Header.Get(kitty.ContentType)

	if strings.HasPrefix(header, kitty.MultipartFormData) {
		return s.Multipart.Form.Empty(key)
	}

	if strings.HasPrefix(header, kitty.ApplicationFormUrlencoded) {
		return s.Form.Empty(key)
	}

	if strings.HasPrefix(header, kitty.ApplicationJson) {
		return s.Json.Empty(key)
	}

	return false
}

func (s *Stream) AutoGet(key string) Value {
	if strings.ToUpper(s.Request.Method) == "GET" {
		return s.Query.First(key)
	}

	var header = s.Request.Header.Get(kitty.ContentType)

	if strings.HasPrefix(header, kitty.MultipartFormData) {
		return s.Multipart.Form.First(key)
	}

	if strings.HasPrefix(header, kitty.ApplicationFormUrlencoded) {
		return s.Form.First(key)
	}

	if strings.HasPrefix(header, kitty.ApplicationJson) {
		return s.Json.Get(key)
	}

	return Value{}
}

func (s *Stream) Url() string {
	var buf bytes.Buffer
	var host = s.Host()
	buf.WriteString(s.Scheme() + "://" + host + s.Request.URL.Path)
	if s.Request.URL.RawQuery != "" {
		buf.WriteString("?" + s.Request.URL.RawQuery)
	}
	if s.Request.URL.Fragment != "" {
		buf.WriteString("#" + s.Request.URL.Fragment)
	}
	return buf.String()
}

func (s *Stream) String() string {

	var header = s.Request.Header.Get(kitty.ContentType)

	if strings.ToUpper(s.Request.Method) == "GET" {
		return s.Query.String()
	}

	if strings.HasPrefix(header, kitty.MultipartFormData) {
		var filesStr = s.Multipart.Files.String()
		var formStr = s.Multipart.Form.String()
		if filesStr == "" {
			return formStr
		}
		return formStr + " " + filesStr
	}

	if strings.HasPrefix(header, kitty.ApplicationFormUrlencoded) {
		return s.Form.String()
	}

	if strings.HasPrefix(header, kitty.ApplicationJson) {
		return s.Json.String()
	}

	if strings.HasPrefix(header, kitty.ApplicationProtobuf) {
		return "<Protobuf: " + strconv.Itoa(len(s.Protobuf.Bytes())) + " >"
	}

	return ""
}

func (s *Stream) Scheme() string {
	var scheme = "http"
	if s.Request.TLS != nil {
		scheme = "https"
	}
	return scheme
}
