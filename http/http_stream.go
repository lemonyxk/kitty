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

	"github.com/lemoyxk/kitty"
)

func NewStream(w http.ResponseWriter, r *http.Request) *Stream {
	return &Stream{Response: w, Request: r}
}

type Stream struct {
	// Server   *Server
	Response http.ResponseWriter
	Request  *http.Request
	Query    *Store
	Form     *Store
	Json     *Json
	Files    *Files

	Params  kitty.Params
	Context kitty.Context
	Logger  kitty.Logger

	maxMemory     int64
	hasParseQuery bool
	hasParseForm  bool
	hasParseJson  bool
	hasParseFiles bool
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
	s.SetHeader("Content-Type", "application/json")
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
	s.SetHeader("Content-Type", "application/octet-stream")
	s.SetHeader("content-Disposition", "attachment;filename="+fileName)
	return s.End(content)
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

func (s *Stream) ParseJson() *Json {

	if s.hasParseJson {
		return s.Json
	}

	s.hasParseJson = true

	s.Json = &Json{}

	jsonBody, err := ioutil.ReadAll(s.Request.Body)
	if err != nil {
		return s.Json
	}

	s.Json.any = jsoniter.Get(jsonBody)

	return s.Json
}

func (s *Stream) ParseFiles() *Files {

	if s.hasParseFiles {
		return s.Files
	}

	s.hasParseFiles = true

	s.Files = &Files{}

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

	s.Form = &Store{}

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

	s.Query = &Store{}

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

	s.Form = &Store{}

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

	var header = s.Request.Header.Get("Content-Type")

	if strings.ToUpper(s.Request.Method) == "GET" {
		s.ParseQuery()
		return
	}

	if strings.HasPrefix(header, "multipart/form-data") {
		s.ParseMultipart()
		s.ParseFiles()
		return
	}

	if strings.HasPrefix(header, "application/x-www-form-urlencoded") {
		s.ParseForm()
		return
	}

	if strings.HasPrefix(header, "application/json") {
		s.ParseJson()
		return
	}
}

func (s *Stream) AutoGet(key string) Value {
	if strings.ToUpper(s.Request.Method) == "GET" {
		return s.Query.First(key)
	}

	var header = s.Request.Header.Get("Content-Type")

	if strings.HasPrefix(header, "multipart/form-data") {
		return s.Form.First(key)
	}

	if strings.HasPrefix(header, "application/x-www-form-urlencoded") {
		return s.Form.First(key)
	}

	if strings.HasPrefix(header, "application/json") {
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

	var header = s.Request.Header.Get("Content-Type")

	if strings.ToUpper(s.Request.Method) == "GET" {
		return s.Query.String()
	}

	if strings.HasPrefix(header, "multipart/form-data") {
		return s.Form.String()
	}

	if strings.HasPrefix(header, "application/x-www-form-urlencoded") {
		return s.Form.String()
	}

	if strings.HasPrefix(header, "application/json") {
		return s.Json.String()
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
