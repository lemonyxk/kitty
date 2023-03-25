package http

import (
	"bytes"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/lemonyxk/kitty/kitty"
	"github.com/lemonyxk/kitty/kitty/header"
	"github.com/lemonyxk/kitty/socket"
)

type Stream struct {
	// Server   *Server
	Response http.ResponseWriter
	Request  *http.Request

	Query    *Store
	Form     *Store
	Files    *Files
	Json     *Json
	Protobuf *Protobuf

	Params  socket.Params
	Context kitty.Context
	Logger  kitty.Logger

	Sender *Sender

	Parser *Parser
}

func NewStream(w http.ResponseWriter, r *http.Request) *Stream {
	var stream = &Stream{
		Response: w, Request: r,
		Protobuf: &Protobuf{},
		Query:    &Store{},
		Form:     &Store{},
		Json:     &Json{},
		Files:    &Files{},
		Sender:   &Sender{response: w, request: r},
		Parser:   &Parser{response: w, request: r, maxMemory: 6 * 1024 * 1024},
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
	if host := s.Request.Header.Get(header.Host); host != "" {
		return host
	}
	return s.Request.Host
}

func (s *Stream) ClientIP() string {

	if ip := strings.Split(s.Request.Header.Get(header.XForwardedFor), ",")[0]; ip != "" {
		return ip
	}

	if ip := s.Request.Header.Get(header.XRealIP); ip != "" {
		return ip
	}

	if ip, _, err := net.SplitHostPort(s.Request.RemoteAddr); err == nil {
		return ip
	}

	return ""
}

func (s *Stream) Has(key string) bool {
	if !s.Parser.HasParse() {
		return false
	}

	if strings.ToUpper(s.Request.Method) == http.MethodGet {
		return s.Query.Has(key)
	}

	if strings.ToUpper(s.Request.Method) == http.MethodHead {
		return s.Query.Has(key)
	}

	if strings.ToUpper(s.Request.Method) == http.MethodTrace {
		return s.Query.Has(key)
	}

	// May have a request body
	if strings.ToUpper(s.Request.Method) == http.MethodDelete {
		return s.Form.Has(key)
	}

	if strings.ToUpper(s.Request.Method) == http.MethodOptions {
		return s.Query.Has(key)
	}

	if strings.ToUpper(s.Request.Method) == http.MethodConnect {
		return s.Query.Has(key)
	}

	var contentType = s.Request.Header.Get(header.ContentType)

	if strings.HasPrefix(contentType, header.MultipartFormData) {
		return s.Form.Has(key) || s.Files.Has(key)
	}

	if strings.HasPrefix(contentType, header.ApplicationFormUrlencoded) {
		return s.Form.Has(key)
	}

	if strings.HasPrefix(contentType, header.ApplicationJson) {
		return s.Json.Has(key)
	}

	// Don't support protobuf
	// if strings.HasPrefix(header, kitty.ApplicationProtobuf) {
	//
	// }

	return false
}

func (s *Stream) Empty(key string) bool {

	if !s.Parser.HasParse() {
		return false
	}

	if strings.ToUpper(s.Request.Method) == http.MethodGet {
		return s.Query.Empty(key)
	}

	if strings.ToUpper(s.Request.Method) == http.MethodHead {
		return s.Query.Empty(key)
	}

	if strings.ToUpper(s.Request.Method) == http.MethodTrace {
		return s.Query.Empty(key)
	}

	// May have a request body
	if strings.ToUpper(s.Request.Method) == http.MethodDelete {
		return s.Form.Empty(key)
	}

	if strings.ToUpper(s.Request.Method) == http.MethodOptions {
		return s.Query.Empty(key)
	}

	if strings.ToUpper(s.Request.Method) == http.MethodConnect {
		return s.Query.Empty(key)
	}

	var contentType = s.Request.Header.Get(header.ContentType)

	if strings.HasPrefix(contentType, header.MultipartFormData) {
		return s.Form.Empty(key) || s.Files.Empty(key)
	}

	if strings.HasPrefix(contentType, header.ApplicationFormUrlencoded) {
		return s.Form.Empty(key)
	}

	if strings.HasPrefix(contentType, header.ApplicationJson) {
		return s.Json.Empty(key)
	}

	// Don't support protobuf
	// if strings.HasPrefix(header, kitty.ApplicationProtobuf) {
	//
	// }

	return false
}

func (s *Stream) AutoGet(key string) Value {
	if !s.Parser.HasParse() {
		return Value{}
	}

	if strings.ToUpper(s.Request.Method) == http.MethodGet {
		return s.Query.First(key)
	}

	if strings.ToUpper(s.Request.Method) == http.MethodHead {
		return s.Query.First(key)
	}

	if strings.ToUpper(s.Request.Method) == http.MethodTrace {
		return s.Query.First(key)
	}

	// May have a request body
	if strings.ToUpper(s.Request.Method) == http.MethodDelete {
		return s.Form.First(key)
	}

	if strings.ToUpper(s.Request.Method) == http.MethodOptions {
		return s.Query.First(key)
	}

	if strings.ToUpper(s.Request.Method) == http.MethodConnect {
		return s.Query.First(key)
	}

	var contentType = s.Request.Header.Get(header.ContentType)

	if strings.HasPrefix(contentType, header.MultipartFormData) {
		var res = s.Form.First(key)
		if res.v != nil && *res.v != "" {
			return res
		}
		return s.Files.Name(key)
	}

	if strings.HasPrefix(contentType, header.ApplicationFormUrlencoded) {
		return s.Form.First(key)
	}

	if strings.HasPrefix(contentType, header.ApplicationJson) {
		return s.Json.Get(key)
	}

	// Don't support protobuf
	// if strings.HasPrefix(header, kitty.ApplicationProtobuf) {
	//
	// }

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

	var contentType = s.Request.Header.Get(header.ContentType)

	if strings.ToUpper(s.Request.Method) == "GET" {
		return s.Query.String()
	}

	if strings.HasPrefix(contentType, header.MultipartFormData) {
		return strings.Join([]string{s.Form.String(), s.Files.String()}, " ")
	}

	if strings.HasPrefix(contentType, header.ApplicationFormUrlencoded) {
		return s.Form.String()
	}

	if strings.HasPrefix(contentType, header.ApplicationJson) {
		return s.Json.String()
	}

	if strings.HasPrefix(contentType, header.ApplicationProtobuf) {
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
