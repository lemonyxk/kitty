/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2023-02-27 23:47
**/

package http

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/json-iterator/go"
	"github.com/lemonyxk/kitty/v2/kitty"
)

type Parser struct {
	stream *Stream

	response http.ResponseWriter
	request  *http.Request

	maxMemory         int64
	hasParseQuery     bool
	hasParseForm      bool
	hasParseMultipart bool
	hasParseJson      bool
	hasParseProtobuf  bool
}

func (s *Parser) SetMaxMemory(maxMemory int64) {
	s.maxMemory = maxMemory
}

func (s *Parser) Json() *Json {

	if s.hasParseJson {
		return s.stream.Json
	}

	s.hasParseJson = true

	jsonBody, err := io.ReadAll(s.request.Body)
	if err != nil {
		return s.stream.Json
	}

	s.stream.Json.any = jsoniter.Get(jsonBody)
	s.stream.Json.bts = jsonBody

	return s.stream.Json
}

func (s *Parser) Protobuf() *Protobuf {

	if s.hasParseProtobuf {
		return s.stream.Protobuf
	}

	s.hasParseProtobuf = true

	protobufBody, err := io.ReadAll(s.request.Body)
	if err != nil {
		return s.stream.Protobuf
	}

	s.stream.Protobuf.bts = protobufBody

	return s.stream.Protobuf
}

func (s *Parser) Multipart() *Multipart {

	if s.hasParseMultipart {
		return s.stream.Multipart
	}

	s.hasParseMultipart = true

	err := s.request.ParseMultipartForm(s.maxMemory)
	if err != nil {
		return s.stream.Multipart
	}

	var parse = s.request.MultipartForm.Value

	for k, v := range parse {
		s.stream.Multipart.Form.keys = append(s.stream.Multipart.Form.keys, k)
		s.stream.Multipart.Form.values = append(s.stream.Multipart.Form.values, v)
	}

	var data = s.request.MultipartForm.File

	s.stream.Multipart.Files.files = data

	return s.stream.Multipart
}

func (s *Parser) Query() *Store {

	if s.hasParseQuery {
		return s.stream.Query
	}

	s.hasParseQuery = true

	var params = s.request.URL.RawQuery

	parse, err := url.ParseQuery(params)
	if err != nil {
		return s.stream.Query
	}

	for k, v := range parse {
		s.stream.Query.keys = append(s.stream.Query.keys, k)
		s.stream.Query.values = append(s.stream.Query.values, v)
	}

	return s.stream.Query
}

func (s *Parser) Form() *Store {

	if s.hasParseForm {
		return s.stream.Form
	}

	s.hasParseForm = true

	err := s.request.ParseForm()
	if err != nil {
		return s.stream.Form
	}

	var parse = s.request.PostForm

	for k, v := range parse {
		s.stream.Form.keys = append(s.stream.Form.keys, k)
		s.stream.Form.values = append(s.stream.Form.values, v)
	}

	return s.stream.Form
}

func (s *Parser) Auto() {

	var header = s.request.Header.Get(kitty.ContentType)

	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/GET
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

	// May have a request body
	if strings.ToUpper(s.request.Method) == http.MethodDelete {
		s.Query()
		return
	}

	if strings.ToUpper(s.request.Method) == http.MethodOptions {
		s.Query()
		return
	}

	if strings.HasPrefix(header, kitty.MultipartFormData) {
		s.Multipart()
		return
	}

	if strings.HasPrefix(header, kitty.ApplicationFormUrlencoded) {
		s.Form()
		return
	}

	if strings.HasPrefix(header, kitty.ApplicationJson) {
		s.Json()
		return
	}

	if strings.HasPrefix(header, kitty.ApplicationProtobuf) {
		s.Protobuf()
		return
	}
}
