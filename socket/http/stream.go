package http

import (
	"bytes"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/lemonyxk/kitty/kitty"
	"github.com/lemonyxk/kitty/kitty/header"
	"github.com/lemonyxk/kitty/socket"
)

type Stream[T Packer] struct {
	sender[T]

	//Node *router.Node[*Stream[T]]

	Time time.Time

	Response http.ResponseWriter
	Request  *http.Request

	Query    *Store
	Form     *Store
	File     *File
	Json     *Json
	Protobuf *Protobuf

	Params  socket.Params
	Context kitty.Context
	Logger  kitty.Logger

	Sender *Sender

	Parser *Parser[T]
}

func NewStream[T Packer](conn T, w http.ResponseWriter, r *http.Request) *Stream[T] {
	var stream = &Stream[T]{
		sender:   sender[T]{conn: conn},
		Time:     time.Now(),
		Response: w, Request: r,
		Protobuf: &Protobuf{},
		Query:    &Store{},
		Form:     &Store{},
		Json:     &Json{},
		File:     &File{},
		Sender:   &Sender{response: w, request: r},
		Parser:   &Parser[T]{response: w, request: r, maxMemory: 6 * 1024 * 1024},
	}

	stream.Parser.stream = stream

	return stream
}

func (s *Stream[T]) Forward(fn func(stream *Stream[T]) error) error {
	return fn(s)
}

func (s *Stream[T]) SetHeader(header string, content string) {
	s.Response.Header().Set(header, content)
}

func (s *Stream[T]) Host() string {
	if host := strings.Split(s.Request.Header.Get(header.XForwardedHost), ",")[0]; host != "" {
		return host
	}
	if host := s.Request.Header.Get(header.XRealHost); host != "" {
		return host
	}
	if host := s.Request.Header.Get(header.Host); host != "" {
		return host
	}
	return s.Request.Host
}

func (s *Stream[T]) ClientIP() string {
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

//func (s *Stream[T]) Has(key string) bool {
//	if !s.Parser.HasParse() {
//		return false
//	}
//
//	if strings.ToUpper(s.Request.Method) == http.MethodGet {
//		return s.Query.Has(key)
//	}
//
//	if strings.ToUpper(s.Request.Method) == http.MethodHead {
//		return s.Query.Has(key)
//	}
//
//	if strings.ToUpper(s.Request.Method) == http.MethodTrace {
//		return s.Query.Has(key)
//	}
//
//	// May have a request body
//	if strings.ToUpper(s.Request.Method) == http.MethodDelete {
//		return s.Form.Has(key)
//	}
//
//	if strings.ToUpper(s.Request.Method) == http.MethodOptions {
//		return s.Query.Has(key)
//	}
//
//	if strings.ToUpper(s.Request.Method) == http.MethodConnect {
//		return s.Query.Has(key)
//	}
//
//	var contentType = s.Request.Header.Get(header.ContentType)
//
//	if strings.HasPrefix(contentType, header.MultipartFormData) {
//		return s.Form.Has(key) || s.Files.Has(key)
//	}
//
//	if strings.HasPrefix(contentType, header.ApplicationFormUrlencoded) {
//		return s.Form.Has(key)
//	}
//
//	// Don't support protobuf
//	//if strings.HasPrefix(contentType, header.ApplicationJson) {
//	//	return s.Json.Has(key)
//	//}
//
//	// Don't support protobuf
//	// if strings.HasPrefix(header, kitty.ApplicationProtobuf) {
//	//
//	// }
//
//	return false
//}
//
//func (s *Stream[T]) Empty(key string) bool {
//
//	if !s.Parser.HasParse() {
//		return false
//	}
//
//	if strings.ToUpper(s.Request.Method) == http.MethodGet {
//		return s.Query.Empty(key)
//	}
//
//	if strings.ToUpper(s.Request.Method) == http.MethodHead {
//		return s.Query.Empty(key)
//	}
//
//	if strings.ToUpper(s.Request.Method) == http.MethodTrace {
//		return s.Query.Empty(key)
//	}
//
//	// May have a request body
//	if strings.ToUpper(s.Request.Method) == http.MethodDelete {
//		return s.Form.Empty(key)
//	}
//
//	if strings.ToUpper(s.Request.Method) == http.MethodOptions {
//		return s.Query.Empty(key)
//	}
//
//	if strings.ToUpper(s.Request.Method) == http.MethodConnect {
//		return s.Query.Empty(key)
//	}
//
//	var contentType = s.Request.Header.Get(header.ContentType)
//
//	if strings.HasPrefix(contentType, header.MultipartFormData) {
//		return s.Form.Empty(key) || s.Files.Empty(key)
//	}
//
//	if strings.HasPrefix(contentType, header.ApplicationFormUrlencoded) {
//		return s.Form.Empty(key)
//	}
//
//
//	// Don't support protobuf
//	//if strings.HasPrefix(contentType, header.ApplicationJson) {
//	//	return s.Json.Empty(key)
//	//}
//
//	// Don't support protobuf
//	// if strings.HasPrefix(header, kitty.ApplicationProtobuf) {
//	//
//	// }
//
//	return false
//}

//func (s *Stream[T]) AutoGet(key string) Value {
//	if !s.Parser.HasParse() {
//		return Value{}
//	}
//
//	if strings.ToUpper(s.Request.Method) == http.MethodGet {
//		return s.Query.First(key)
//	}
//
//	if strings.ToUpper(s.Request.Method) == http.MethodHead {
//		return s.Query.First(key)
//	}
//
//	if strings.ToUpper(s.Request.Method) == http.MethodTrace {
//		return s.Query.First(key)
//	}
//
//	// May have a request body
//	if strings.ToUpper(s.Request.Method) == http.MethodDelete {
//		return s.Form.First(key)
//	}
//
//	if strings.ToUpper(s.Request.Method) == http.MethodOptions {
//		return s.Query.First(key)
//	}
//
//	if strings.ToUpper(s.Request.Method) == http.MethodConnect {
//		return s.Query.First(key)
//	}
//
//	var contentType = s.Request.Header.Get(header.ContentType)
//
//	if strings.HasPrefix(contentType, header.MultipartFormData) {
//		var res = s.Form.First(key)
//		if res.v != nil && *res.v != "" {
//			return res
//		}
//		return s.Files.Name(key)
//	}
//
//	if strings.HasPrefix(contentType, header.ApplicationFormUrlencoded) {
//		return s.Form.First(key)
//	}
//
//	// Don't support json
//	//if strings.HasPrefix(contentType, header.ApplicationJson) {
//	//	return s.Json.Get(key)
//	//}
//
//	// Don't support protobuf
//	// if strings.HasPrefix(header, kitty.ApplicationProtobuf) {
//	//
//	// }
//
//	return Value{}
//}

func (s *Stream[T]) Url() string {
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

func (s *Stream[T]) String() string {

	if s.Request.Method == http.MethodGet {
		return s.Query.String()
	}

	var contentType = s.Request.Header.Get(header.ContentType)
	var index = strings.Index(contentType, ";")
	if index > 0 {
		contentType = contentType[:index]
	}

	switch contentType {
	case header.MultipartFormData:
		var res []string
		if len(s.Form.Values) > 0 {
			res = append(res, s.Form.String())
		}
		if len(s.File.FileHeader) > 0 {
			res = append(res, s.File.String())
		}
		return strings.Join(res, " ")
	case header.ApplicationFormUrlencoded:
		return s.Form.String()
	case header.ApplicationJson:
		return s.Json.String()
	case header.ApplicationProtobuf:
		return s.Protobuf.String()
	default:
		return ""
	}
}

func (s *Stream[T]) Scheme() string {
	var scheme = "http"
	if s.Request.TLS != nil {
		scheme = "https"
	}
	return scheme
}

func (s *Stream[T]) UpgradeSse(config *SseConfig) (*Sse[T], error) {
	s.Response.Header().Set(header.ContentType, header.TextEventStream)
	s.Response.Header().Set(header.CacheControl, header.Nocache)
	s.Response.Header().Set(header.Connection, header.KeepAlive)

	s.Response.WriteHeader(http.StatusOK)

	var lastEventId = s.Request.Header.Get(header.LastEventID)

	var id, _ = strconv.ParseInt(lastEventId, 10, 64)

	if config == nil {
		config = &SseConfig{Retry: 3 * time.Second}
	}

	_, err := s.Response.Write([]byte(`retry: ` + strconv.Itoa(int(config.Retry.Milliseconds())) + "\n"))
	if err != nil {
		return nil, err
	}

	var sse = &Sse[T]{Stream: s, LasTEventID: id, close: make(chan struct{}, 1)}

	sse.Flush()

	return sse, nil
}
