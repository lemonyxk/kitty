package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/json-iterator/go"

	"github.com/lemoyxk/kitty"
)

type Files struct {
	files map[string][]*multipart.FileHeader
}

func (f *Files) Files() map[string][]*multipart.FileHeader {
	return f.files
}

func (f *Files) First(fileName string) *multipart.FileHeader {
	if file, ok := f.files[fileName]; ok {
		return file[0]
	}
	return nil
}

func (f *Files) Index(fileName string, index int) *multipart.FileHeader {
	if file, ok := f.files[fileName]; ok {
		return file[index]
	}
	return nil
}

func (f *Files) All(fileName string) []*multipart.FileHeader {
	if file, ok := f.files[fileName]; ok {
		return file
	}
	return nil
}

type Value struct {
	v *string
}

func (v Value) Int() int {
	if v.v == nil {
		return 0
	}
	r, _ := strconv.Atoi(*v.v)
	return r
}

func (v Value) Float64() float64 {
	if v.v == nil {
		return 0
	}
	r, _ := strconv.ParseFloat(*v.v, 64)
	return r
}

func (v Value) String() string {
	if v.v == nil {
		return ""
	}
	return *v.v
}

func (v Value) Bool() bool {
	return strings.ToUpper(v.String()) == "TRUE"
}

func (v Value) Bytes() []byte {
	return []byte(v.String())
}

type Json struct {
	any jsoniter.Any
}

func (j *Json) Reset(data interface{}) jsoniter.Any {
	bts, _ := jsoniter.Marshal(data)
	j.any = jsoniter.Get(bts)
	return j.any
}

func (j *Json) getAny() jsoniter.Any {
	if j.any != nil {
		return j.any
	}
	j.any = jsoniter.Get(nil)
	return j.any
}

// GetByID 获取
func (j *Json) Iter() jsoniter.Any {
	return j.getAny()
}

func (j *Json) Has(key string) bool {
	return j.getAny().Get(key).LastError() == nil
}

func (j *Json) Empty(key string) bool {
	return j.getAny().Get(key).ToString() == ""
}

func (j *Json) Get(path ...interface{}) Value {
	var res = j.getAny().Get(path...)
	if res.LastError() != nil {
		return Value{}
	}
	var p = res.ToString()
	return Value{v: &p}
}

func (j *Json) Bytes() []byte {
	return j.Bytes()
}

func (j *Json) String() string {
	return j.getAny().ToString()
}

func (j *Json) Path(path ...interface{}) jsoniter.Any {
	return j.getAny().Get(path...)
}

func (j *Json) Array(path ...interface{}) Array {
	var result []jsoniter.Any
	var val = j.getAny().Get(path...)
	for i := 0; i < val.Size(); i++ {
		result = append(result, val.Get(i))
	}
	return result
}

type Array []jsoniter.Any

func (a Array) String() []string {
	var result []string
	for i := 0; i < len(a); i++ {
		result = append(result, a[i].ToString())
	}
	return result
}

func (a Array) Int() []int {
	var result []int
	for i := 0; i < len(a); i++ {
		result = append(result, a[i].ToInt())
	}
	return result
}

func (a Array) Float64() []float64 {
	var result []float64
	for i := 0; i < len(a); i++ {
		result = append(result, a[i].ToFloat64())
	}
	return result
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

	maxMemory     int64
	hasParseQuery bool
	hasParseForm  bool
	hasParseJson  bool
	hasParseFiles bool
}

func NewStream(w http.ResponseWriter, r *http.Request) *Stream {
	return &Stream{Response: w, Request: r}
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

type JsonFormat struct {
	Status string      `json:"status"`
	Code   int         `json:"code"`
	Msg    interface{} `json:"msg"`
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

	var json = &Json{}

	jsonBody, err := ioutil.ReadAll(s.Request.Body)
	if err != nil {
		return json
	}

	json.any = jsoniter.Get(jsonBody)

	s.Json = json

	return s.Json
}

func (s *Stream) ParseFiles() *Files {

	if s.hasParseFiles {
		return s.Files
	}

	s.hasParseFiles = true

	var file = &Files{}

	err := s.Request.ParseMultipartForm(s.maxMemory)
	if err != nil {
		return file
	}

	var data = s.Request.MultipartForm.File

	file.files = data

	s.Files = file

	return file
}

func (s *Stream) ParseMultipart() *Store {

	if s.hasParseForm {
		return s.Form
	}

	s.hasParseForm = true

	var form = &Store{}

	err := s.Request.ParseMultipartForm(s.maxMemory)
	if err != nil {
		return form
	}

	var parse = s.Request.MultipartForm.Value

	for k, v := range parse {
		form.keys = append(form.keys, k)
		form.values = append(form.values, v)
	}

	s.Form = form

	return form
}

func (s *Stream) ParseQuery() *Store {

	if s.hasParseQuery {
		return s.Query
	}

	s.hasParseQuery = true

	var query = &Store{}

	var params = s.Request.URL.RawQuery

	parse, err := url.ParseQuery(params)
	if err != nil {
		return query
	}

	for k, v := range parse {
		query.keys = append(query.keys, k)
		query.values = append(query.values, v)
	}

	s.Query = query

	return query
}

func (s *Stream) ParseForm() *Store {

	if s.hasParseForm {
		return s.Form
	}

	s.hasParseForm = true

	var form = &Store{}

	err := s.Request.ParseForm()
	if err != nil {
		return form
	}

	var parse = s.Request.PostForm

	for k, v := range parse {
		form.keys = append(form.keys, k)
		form.values = append(form.values, v)
	}

	s.Form = form

	return form
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

type Store struct {
	keys   []string
	values [][]string
}

func (s *Store) Has(key string) bool {
	for i := 0; i < len(s.keys); i++ {
		if s.keys[i] == key {
			return true
		}
	}
	return false
}

func (s *Store) Empty(key string) bool {
	var v = s.First(key).v
	return v == nil || *v == ""
}

func (s *Store) First(key string) Value {
	var res Value
	for i := 0; i < len(s.keys); i++ {
		if s.keys[i] == key {
			res.v = &s.values[i][0]
			return res
		}
	}
	return res
}

func (s *Store) Index(key string, index int) Value {
	var res Value
	for i := 0; i < len(s.keys); i++ {
		if s.keys[i] == key {
			res.v = &s.values[i][index]
			return res
		}
	}
	return res
}

func (s *Store) All(key string) []Value {
	var res []Value
	for i := 0; i < len(s.keys); i++ {
		if s.keys[i] == key {
			for j := 0; j < len(s.values[i]); j++ {
				res = append(res, Value{v: &s.values[i][j]})
			}
		}
	}
	return res
}

func (s *Store) Add(key string, value []string) {
	s.keys = append(s.keys, key)
	s.values = append(s.values, value)
}

func (s *Store) Remove(key string) {
	var index = -1
	for i := 0; i < len(s.keys); i++ {
		if s.keys[i] == key {
			index = i
			break
		}
	}
	if index == -1 {
		return
	}
	s.keys = append(s.keys[0:index], s.keys[index+1:]...)
	s.values = append(s.values[0:index], s.values[index+1:]...)
}

func (s *Store) Keys() []string {
	return s.keys
}

func (s *Store) Values() [][]string {
	return s.values
}

func (s *Store) String() string {

	var buff bytes.Buffer

	for i := 0; i < len(s.keys); i++ {
		buff.WriteString(s.keys[i] + ":")
		for j := 0; j < len(s.values[i]); j++ {
			buff.WriteString(s.values[i][j])
			if j != len(s.values[i])-1 {
				buff.WriteString(",")
			}
		}
		buff.WriteString(" ")
	}

	if buff.Len() == 0 {
		return ""
	}

	var res = buff.String()

	return res[:len(res)-1]
}
