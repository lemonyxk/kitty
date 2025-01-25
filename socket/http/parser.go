/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2023-02-27 23:47
**/

package http

import (
	"bytes"
	"github.com/lemonyxk/kitty/kitty/header"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Parser[T Packer] struct {
	stream *Stream[T]

	response http.ResponseWriter
	request  *http.Request

	maxMemory         int64
	hasParseQuery     bool
	hasParseForm      bool
	hasParseMultipart bool
	hasParseJson      bool
	hasParseProtobuf  bool

	err error
}

//func (s *Parser[T]) HasParse() bool {
//	return s.hasParseQuery || s.hasParseForm || s.hasParseMultipart || s.hasParseJson || s.hasParseProtobuf
//}

func (s *Parser[T]) Error() error {
	return s.err
}

func (s *Parser[T]) SetMaxMemory(maxMemory int64) {
	s.maxMemory = maxMemory
}

func (s *Parser[T]) Json() {

	if s.hasParseJson {
		return
	}

	s.hasParseJson = true

	var buf = new(bytes.Buffer)
	_, err := io.Copy(buf, s.request.Body)
	if err != nil {
		s.err = err
		return
	}

	s.stream.Json.bts = buf.Bytes()

	return
}

func (s *Parser[T]) Protobuf() {

	if s.hasParseProtobuf {
		return
	}

	s.hasParseProtobuf = true

	var buf = new(bytes.Buffer)
	_, err := io.Copy(buf, s.request.Body)
	if err != nil {
		s.err = err
		return
	}

	s.stream.Protobuf.bts = buf.Bytes()

	return
}

func (s *Parser[T]) Multipart() {

	if s.hasParseMultipart {
		return
	}

	s.hasParseMultipart = true

	err := s.request.ParseMultipartForm(s.maxMemory)
	if err != nil {
		s.err = err
		return
	}

	// PostForm includes only post form,
	// Form includes both url query and post form.
	// but only when Content-Type is application/x-www-form-urlencoded
	// and method is POST | PUT | PATCH
	// that url query will be parsed into Form.
	// only when Content-Type is multipart/form-data
	// that file will be parsed into MultipartForm.File

	s.stream.Form.Values = s.request.MultipartForm.Value
	s.stream.File.FileHeader = s.request.MultipartForm.File

	return
}

func (s *Parser[T]) Query() {

	if s.hasParseQuery {
		return
	}

	s.hasParseQuery = true

	var params = s.request.URL.RawQuery

	parse, err := url.ParseQuery(params)
	if err != nil {
		s.err = err
		return
	}

	s.stream.Query.Values = parse

	return
}

func (s *Parser[T]) Form() {

	if s.hasParseForm {
		return
	}

	s.hasParseForm = true

	err := s.request.ParseForm()
	if err != nil {
		s.err = err
		return
	}

	s.stream.Form.Values = s.request.Form

	return
}

func (s *Parser[T]) Auto() {

	switch s.request.Method {
	case http.MethodGet:
		s.Query()
	case http.MethodHead:
		s.Query()
	case http.MethodTrace:
		s.Query()
	case http.MethodConnect:
		s.Query()
	case http.MethodOptions:
		s.Query()
	case http.MethodDelete:
		// May have a request body
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/DELETE

		// cuz DELETE method may have a request body,
		// so we need to parse it.
		// but we need to change the method to post,
		// cuz ParseForm() only support post | put | patch.
		// https://golang.org/pkg/net/http/#Request.ParseForm
		s.request.Method = http.MethodPost
		s.Form()
		s.request.Method = http.MethodDelete
	default:

		var contentType = s.request.Header.Get(header.ContentType)
		var index = strings.Index(contentType, ";")
		if index > 0 {
			contentType = contentType[:index]
		}

		switch contentType {
		case header.MultipartFormData:
			s.Multipart()
		case header.ApplicationFormUrlencoded:
			s.Form()
		case header.ApplicationJson:
			s.Json()
		case header.ApplicationProtobuf:
			s.Protobuf()
		}
	}
}
