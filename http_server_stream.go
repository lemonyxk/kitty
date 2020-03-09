package lemo

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

	"github.com/Lemo-yxk/lemo/exception"
)

type Files struct {
	files map[string][]*multipart.FileHeader
}

func (f Files) All() map[string][]*multipart.FileHeader {
	return f.files
}

func (f Files) Get(fileName string) *multipart.FileHeader {
	if file, ok := f.files[fileName]; ok {
		return file[0]
	}
	return nil
}

type Params struct {
	Keys   []string
	Values []string
}

func (ps Params) ByName(name string) string {
	for i := 0; i < len(ps.Keys); i++ {
		if ps.Keys[i] == name {
			return ps.Values[i]
		}
	}
	return ""
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

func (j Json) Reset(data interface{}) jsoniter.Any {
	bts, _ := jsoniter.Marshal(data)
	j.any = jsoniter.Get(bts)
	return j.any
}

func (j Json) getAny() jsoniter.Any {
	if j.any != nil {
		return j.any
	}
	j.any = jsoniter.Get(nil)
	return j.any
}

// GetByID 获取
func (j Json) Iter() jsoniter.Any {
	return j.getAny()
}

func (j Json) Has(key string) bool {
	return j.getAny().Get(key).LastError() == nil
}

func (j Json) Empty(key string) bool {
	return j.getAny().Get(key).ToString() == ""
}

func (j Json) Get(path ...interface{}) Value {
	var res = j.getAny().Get(path...)
	if res.LastError() != nil {
		return Value{}
	}
	var p = res.ToString()
	return Value{v: &p}
}

func (j Json) Bytes() []byte {
	return j.Bytes()
}

func (j Json) String() string {
	return j.getAny().ToString()
}

func (j Json) Path(path ...interface{}) jsoniter.Any {
	return j.getAny().Get(path...)
}

func (j Json) Array(path ...interface{}) Array {
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
	Server   *HttpServer
	Response http.ResponseWriter
	Request  *http.Request
	Params   Params
	Context  Context
	Query    Store
	Form     Store
	Json     Json
	Files    Files

	maxMemory     int64
	error         interface{}
	hasParseQuery bool
	hasParseForm  bool
	hasParseJson  bool
	hasParseFiles bool
}

func NewStream(h *HttpServer, w http.ResponseWriter, r *http.Request) *Stream {
	return &Stream{Server: h, Response: w, Request: r}
}

func (stream *Stream) LastError() interface{} {
	return stream.error
}

func (stream *Stream) Match() bool {
	return stream.error != nil
}

func (stream *Stream) Forward(fn HttpServerFunction) exception.ErrorFunc {
	return fn(stream)
}

func (stream *Stream) SetMaxMemory(maxMemory int64) {
	stream.maxMemory = maxMemory
}

func (stream *Stream) SetHeader(header string, content string) {
	stream.Response.Header().Set(header, content)
}

func (stream *Stream) JsonFormat(status string, code int, msg interface{}) exception.ErrorFunc {
	return exception.New(stream.EndJson(JsonFormat{status, code, msg}))
}

func (stream *Stream) End(data interface{}) error {
	var err error
	switch data.(type) {
	case []byte:
		err = stream.EndBytes(data.([]byte))
	case string:
		err = stream.EndString(data.(string))
	default:
		err = stream.EndString(fmt.Sprintf("%v", data))
	}
	return err
}

func (stream *Stream) EndJson(data interface{}) error {
	stream.SetHeader("Content-Type", "application/json")
	var bts, err = jsoniter.Marshal(data)
	_, err = stream.Response.Write(bts)
	return err
}

func (stream *Stream) EndString(data string) error {
	_, err := stream.Response.Write([]byte(data))
	return err
}

func (stream *Stream) EndBytes(data []byte) error {
	_, err := stream.Response.Write(data)
	return err
}

func (stream *Stream) EndFile(fileName string, content []byte) error {
	stream.SetHeader("Content-Type", "application/octet-stream")
	stream.SetHeader("content-Disposition", "attachment;filename="+fileName)
	return stream.EndBytes(content)
}

func (stream *Stream) Host() string {
	if host := stream.Request.Header.Get(Host); host != "" {
		return host
	}
	return stream.Request.Host
}

func (stream *Stream) ClientIP() string {

	if ip := strings.Split(stream.Request.Header.Get(XForwardedFor), ",")[0]; ip != "" {
		return ip
	}

	if ip := stream.Request.Header.Get(XRealIP); ip != "" {
		return ip
	}

	if ip, _, err := net.SplitHostPort(stream.Request.RemoteAddr); err == nil {
		return ip
	}

	return ""
}

func (stream *Stream) ParseJson() Json {

	if stream.hasParseJson {
		return stream.Json
	}

	stream.hasParseJson = true

	var json = Json{}

	jsonBody, err := ioutil.ReadAll(stream.Request.Body)
	if err != nil {
		return json
	}

	json.any = jsoniter.Get(jsonBody)

	stream.Json = json

	return stream.Json
}

func (stream *Stream) ParseFiles() Files {

	if stream.hasParseFiles {
		return stream.Files
	}

	stream.hasParseFiles = true

	var file = Files{}

	err := stream.Request.ParseMultipartForm(stream.maxMemory)
	if err != nil {
		return file
	}

	var data = stream.Request.MultipartForm.File

	file.files = data

	stream.Files = file

	return file
}

func (stream *Stream) ParseMultipart() Store {

	if stream.hasParseForm {
		return stream.Form
	}

	stream.hasParseForm = true

	var form = Store{}

	err := stream.Request.ParseMultipartForm(stream.maxMemory)
	if err != nil {
		return form
	}

	var parse = stream.Request.MultipartForm.Value

	for k, v := range parse {
		form.keys = append(form.keys, k)
		form.values = append(form.values, v[0])
	}

	stream.Form = form

	return form
}

func (stream *Stream) ParseQuery() Store {

	if stream.hasParseQuery {
		return stream.Query
	}

	stream.hasParseQuery = true

	var query = Store{}

	var params = stream.Request.URL.RawQuery

	parse, err := url.ParseQuery(params)
	if err != nil {
		return query
	}

	for k, v := range parse {
		query.keys = append(query.keys, k)
		query.values = append(query.values, v[0])
	}

	stream.Query = query

	return query
}

func (stream *Stream) ParseForm() Store {

	if stream.hasParseForm {
		return stream.Form
	}

	stream.hasParseForm = true

	var form = Store{}

	err := stream.Request.ParseForm()
	if err != nil {
		return form
	}

	var parse = stream.Request.PostForm

	for k, v := range parse {
		form.keys = append(form.keys, k)
		form.values = append(form.values, v[0])
	}

	stream.Form = form

	return form
}

func (stream *Stream) AutoParse() {

	var header = stream.Request.Header.Get("Content-Type")

	if strings.ToUpper(stream.Request.Method) == "GET" {
		stream.ParseQuery()
		return
	}

	if strings.HasPrefix(header, "multipart/form-data") {
		stream.ParseMultipart()
		stream.ParseFiles()
		return
	}

	if strings.HasPrefix(header, "application/x-www-form-urlencoded") {
		stream.ParseForm()
		return
	}

	if strings.HasPrefix(header, "application/json") {
		stream.ParseJson()
		return
	}
}

func (stream *Stream) AutoGet(key string) Value {
	if strings.ToUpper(stream.Request.Method) == "GET" {
		return stream.Query.Get(key)
	}

	var header = stream.Request.Header.Get("Content-Type")

	if strings.HasPrefix(header, "multipart/form-data") {
		return stream.Form.Get(key)
	}

	if strings.HasPrefix(header, "application/x-www-form-urlencoded") {
		return stream.Form.Get(key)
	}

	if strings.HasPrefix(header, "application/json") {
		return stream.Json.Get(key)
	}

	return Value{}
}

func (stream *Stream) Url() string {
	var buf bytes.Buffer
	var host = stream.Host()
	buf.WriteString(stream.Scheme() + "://" + host + stream.Request.URL.Path)
	if stream.Request.URL.RawQuery != "" {
		buf.WriteString("?" + stream.Request.URL.RawQuery)
	}
	if stream.Request.URL.Fragment != "" {
		buf.WriteString("#" + stream.Request.URL.Fragment)
	}
	return buf.String()
}

func (stream *Stream) String() string {

	var header = stream.Request.Header.Get("Content-Type")

	if strings.ToUpper(stream.Request.Method) == "GET" {
		return stream.Query.String()
	}

	if strings.HasPrefix(header, "multipart/form-data") {
		return stream.Form.String()
	}

	if strings.HasPrefix(header, "application/x-www-form-urlencoded") {
		return stream.Form.String()
	}

	if strings.HasPrefix(header, "application/json") {
		return stream.Json.String()
	}

	return ""
}

func (stream *Stream) Scheme() string {
	var scheme = "http"
	if stream.Request.TLS != nil {
		scheme = "https"
	}
	return scheme
}

type Store struct {
	keys   []string
	values []string
}

func (store Store) Has(key string) bool {
	for i := 0; i < len(store.keys); i++ {
		if store.keys[i] == key {
			return true
		}
	}
	return false
}

func (store Store) Empty(key string) bool {
	var v = store.Get(key).v
	return v == nil || *v == ""
}

func (store Store) Get(key string) Value {
	var val = Value{}
	for i := 0; i < len(store.keys); i++ {
		if store.keys[i] == key {
			val.v = &store.values[i]
			return val
		}
	}
	return val
}

func (store Store) Add(key string, value string) {
	store.keys = append(store.keys, key)
	store.values = append(store.values, value)
}

func (store Store) Remove(key string) {
	var index = -1
	for i := 0; i < len(store.keys); i++ {
		if store.keys[i] == key {
			index = i
			break
		}
	}
	if index == -1 {
		return
	}
	store.keys = append(store.keys[0:index], store.keys[index+1:]...)
	store.values = append(store.values[0:index], store.values[index+1:]...)
}

func (store Store) Keys() []string {
	return store.keys
}

func (store Store) Values() []string {
	return store.values
}

func (store Store) String() string {

	var buff bytes.Buffer

	for i := 0; i < len(store.keys); i++ {
		buff.WriteString(store.keys[i] + ":")
		buff.WriteString(store.values[i])
		buff.WriteString(" ")
	}

	if buff.Len() == 0 {
		return ""
	}

	var res = buff.String()

	return res[:len(res)-1]
}
