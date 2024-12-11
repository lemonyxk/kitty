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
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/lemonyxk/kitty/kitty/header"
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

	s.stream.Json.buf = buf

	return
}

func (s *Parser[T]) Protobuf() {

	if s.hasParseProtobuf {
		return
	}

	s.hasParseProtobuf = true

	protobufBody, err := io.ReadAll(s.request.Body)
	if err != nil {
		s.err = err
		return
	}

	s.stream.Protobuf.bts = protobufBody

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
	var parse = s.request.Form

	for k, v := range parse {
		s.stream.Form.keys = append(s.stream.Form.keys, k)
		s.stream.Form.values = append(s.stream.Form.values, v)
	}

	s.stream.Files.files = s.request.MultipartForm.File

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

	for k, v := range parse {
		s.stream.Query.keys = append(s.stream.Query.keys, k)
		s.stream.Query.values = append(s.stream.Query.values, v)
	}

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

	var parse = s.request.Form

	for k, v := range parse {
		s.stream.Form.keys = append(s.stream.Form.keys, k)
		s.stream.Form.values = append(s.stream.Form.values, v)
	}

	return
}

func (s *Parser[T]) Auto() {
	if strings.ToUpper(s.request.Method) == http.MethodGet {
		s.Query()
		return
	}

	if strings.ToUpper(s.request.Method) == http.MethodHead {
		s.Query()
		return
	}

	if strings.ToUpper(s.request.Method) == http.MethodTrace {
		s.Query()
		return
	}

	if strings.ToUpper(s.request.Method) == http.MethodConnect {
		s.Query()
		return
	}

	// May have a request body
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/DELETE
	if strings.ToUpper(s.request.Method) == http.MethodDelete {
		// cuz DELETE method may have a request body,
		// so we need to parse it.
		// but we need to change the method to post,
		// cuz ParseForm() only support post | put | patch.
		// https://golang.org/pkg/net/http/#Request.ParseForm
		s.request.Method = http.MethodPost
		s.Form()
		s.request.Method = http.MethodDelete
		return
	}

	if strings.ToUpper(s.request.Method) == http.MethodOptions {
		s.Query()
		return
	}

	var contentType = s.request.Header.Get(header.ContentType)

	if strings.HasPrefix(contentType, header.MultipartFormData) {
		s.Multipart()
		return
	}

	if strings.HasPrefix(contentType, header.ApplicationFormUrlencoded) {
		s.Form()
		return
	}

	if strings.HasPrefix(contentType, header.ApplicationJson) {
		s.Json()
		return
	}

	if strings.HasPrefix(contentType, header.ApplicationProtobuf) {
		s.Protobuf()
		return
	}
}
