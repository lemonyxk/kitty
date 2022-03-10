package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/json-iterator/go"
	kitty2 "github.com/lemoyxk/kitty/v2/kitty"
)

type Stream struct {
	// Server   *Server
	Response http.ResponseWriter
	Request  *http.Request
	Query    *Store
	Form     *Store
	Json     *Json
	Protobuf *Protobuf
	Files    *Files

	Params  kitty2.Params
	Context kitty2.Context
	Logger  kitty2.Logger

	maxMemory        int64
	hasParseQuery    bool
	hasParseForm     bool
	hasParseJson     bool
	hasParseProtobuf bool
	hasParseFiles    bool
}

func (s *Stream) Forward(fn func(stream *Stream) error) error {
	return fn(s)
}

func (s *Stream) SetMaxMemory(maxMemory int64) {
	s.maxMemory = maxMemory
}

func (s *Stream) SetHeader(header string, content string) {
	s.Response.Header().Set(header, content)
}

func (s *Stream) JsonFormat(status string, code int, msg interface{}) error {
	return s.EndJson(JsonFormat{Status: status, Code: code, Msg: msg})
}

func (s *Stream) End(data interface{}) error {
	switch data.(type) {
	case []byte:
		return s.EndBytes(data.([]byte))
	case string:
		return s.EndString(data.(string))
	default:
		return s.EndString(fmt.Sprintf("%v", data))
	}
}

func (s *Stream) EndJson(data interface{}) error {
	s.SetHeader(kitty2.ContentType, kitty2.ApplicationJson)
	bts, err := jsoniter.Marshal(data)
	if err != nil {
		return err
	}
	_, err = s.Response.Write(bts)
	return err
}

func (s *Stream) EndString(data string) error {
	_, err := s.Response.Write([]byte(data))
	return err
}

func (s *Stream) EndBytes(data []byte) error {
	_, err := s.Response.Write(data)
	return err
}

func (s *Stream) EndFile(fileName string, content interface{}) error {
	s.SetHeader(kitty2.ContentType, kitty2.ApplicationOctetStream)
	s.SetHeader("content-Disposition", "attachment;filename="+fileName)
	return s.End(content)
}

func (s *Stream) Host() string {
	if host := s.Request.Header.Get(kitty2.Host); host != "" {
		return host
	}
	return s.Request.Host
}

func (s *Stream) ClientIP() string {

	if ip := strings.Split(s.Request.Header.Get(kitty2.XForwardedFor), ",")[0]; ip != "" {
		return ip
	}

	if ip := s.Request.Header.Get(kitty2.XRealIP); ip != "" {
		return ip
	}

	if ip, _, err := net.SplitHostPort(s.Request.RemoteAddr); err == nil {
		return ip
	}

	return ""
}

func (s *Stream) ParseJson() *Json {

	if s.hasParseJson {
		return s.Json
	}

	s.hasParseJson = true

	jsonBody, err := ioutil.ReadAll(s.Request.Body)
	if err != nil {
		return s.Json
	}

	s.Json.any = jsoniter.Get(jsonBody)
	s.Json.bts = jsonBody

	return s.Json
}

func (s *Stream) ParseProtobuf() *Protobuf {

	if s.hasParseProtobuf {
		return s.Protobuf
	}

	s.hasParseProtobuf = true

	protobufBody, err := ioutil.ReadAll(s.Request.Body)
	if err != nil {
		return s.Protobuf
	}

	s.Protobuf.bts = protobufBody

	return s.Protobuf
}

func (s *Stream) ParseFiles() *Files {

	if s.hasParseFiles {
		return s.Files
	}

	s.hasParseFiles = true

	err := s.Request.ParseMultipartForm(s.maxMemory)
	if err != nil {
		return s.Files
	}

	var data = s.Request.MultipartForm.File

	s.Files.files = data

	return s.Files
}

func (s *Stream) ParseMultipart() *Store {

	if s.hasParseForm {
		return s.Form
	}

	s.hasParseForm = true

	err := s.Request.ParseMultipartForm(s.maxMemory)
	if err != nil {
		return s.Form
	}

	var parse = s.Request.MultipartForm.Value

	for k, v := range parse {
		s.Form.keys = append(s.Form.keys, k)
		s.Form.values = append(s.Form.values, v)
	}

	return s.Form
}

func (s *Stream) ParseQuery() *Store {

	if s.hasParseQuery {
		return s.Query
	}

	s.hasParseQuery = true

	var params = s.Request.URL.RawQuery

	parse, err := url.ParseQuery(params)
	if err != nil {
		return s.Query
	}

	for k, v := range parse {
		s.Query.keys = append(s.Query.keys, k)
		s.Query.values = append(s.Query.values, v)
	}

	return s.Query
}

func (s *Stream) ParseForm() *Store {

	if s.hasParseForm {
		return s.Form
	}

	s.hasParseForm = true

	err := s.Request.ParseForm()
	if err != nil {
		return s.Form
	}

	var parse = s.Request.PostForm

	for k, v := range parse {
		s.Form.keys = append(s.Form.keys, k)
		s.Form.values = append(s.Form.values, v)
	}

	return s.Form
}

func (s *Stream) AutoParse() {

	var header = s.Request.Header.Get(kitty2.ContentType)

	if strings.ToUpper(s.Request.Method) == "GET" {
		s.ParseQuery()
		return
	}

	if strings.HasPrefix(header, kitty2.MultipartFormData) {
		s.ParseMultipart()
		s.ParseFiles()
		return
	}

	if strings.HasPrefix(header, kitty2.ApplicationFormUrlencoded) {
		s.ParseForm()
		return
	}

	if strings.HasPrefix(header, kitty2.ApplicationJson) {
		s.ParseJson()
		return
	}

	if strings.HasPrefix(header, kitty2.ApplicationProtobuf) {
		s.ParseProtobuf()
		return
	}
}

func (s *Stream) Has(key string) bool {
	if strings.ToUpper(s.Request.Method) == "GET" {
		return s.Query.Has(key)
	}

	var header = s.Request.Header.Get(kitty2.ContentType)

	if strings.HasPrefix(header, kitty2.MultipartFormData) {
		return s.Form.Has(key)
	}

	if strings.HasPrefix(header, kitty2.ApplicationFormUrlencoded) {
		return s.Form.Has(key)
	}

	if strings.HasPrefix(header, kitty2.ApplicationJson) {
		return s.Json.Has(key)
	}

	return false
}

func (s *Stream) Empty(key string) bool {
	if strings.ToUpper(s.Request.Method) == "GET" {
		return s.Query.Empty(key)
	}

	var header = s.Request.Header.Get(kitty2.ContentType)

	if strings.HasPrefix(header, kitty2.MultipartFormData) {
		return s.Form.Empty(key)
	}

	if strings.HasPrefix(header, kitty2.ApplicationFormUrlencoded) {
		return s.Form.Empty(key)
	}

	if strings.HasPrefix(header, kitty2.ApplicationJson) {
		return s.Json.Empty(key)
	}

	return false
}

func (s *Stream) AutoGet(key string) Value {
	if strings.ToUpper(s.Request.Method) == "GET" {
		return s.Query.First(key)
	}

	var header = s.Request.Header.Get(kitty2.ContentType)

	if strings.HasPrefix(header, kitty2.MultipartFormData) {
		return s.Form.First(key)
	}

	if strings.HasPrefix(header, kitty2.ApplicationFormUrlencoded) {
		return s.Form.First(key)
	}

	if strings.HasPrefix(header, kitty2.ApplicationJson) {
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

	var header = s.Request.Header.Get(kitty2.ContentType)

	if strings.ToUpper(s.Request.Method) == "GET" {
		return s.Query.String()
	}

	if strings.HasPrefix(header, kitty2.MultipartFormData) {
		return s.Form.String()
	}

	if strings.HasPrefix(header, kitty2.ApplicationFormUrlencoded) {
		return s.Form.String()
	}

	if strings.HasPrefix(header, kitty2.ApplicationJson) {
		return s.Json.String()
	}

	if strings.HasPrefix(header, kitty2.ApplicationProtobuf) {
		return string(s.Protobuf.Bytes())
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
