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

func (f *Files) All() map[string][]*multipart.FileHeader {
	return f.files
}

func (f *Files) Get(fileName string) []*multipart.FileHeader {
	if file, ok := f.files[fileName]; ok {
		return file
	}
	return nil
}

type Params struct {
	Keys   []string
	Values []string
}

func (ps *Params) ByName(name string) string {
	for i := 0; i < len(ps.Keys); i++ {
		if ps.Keys[i] == name {
			return ps.Values[i]
		}
	}
	return ""
}

type Value struct {
	v interface{}
}

func (v *Value) Int() int {

	switch v.v.(type) {
	case int:
		return v.v.(int)
	case string:
		r, err := strconv.Atoi(v.v.(string))
		if err != nil {
			return 0
		}
		return r
	case float64:
		return int(v.v.(float64))
	default:
		return 0
	}
}

func (v *Value) Float64() float64 {
	switch v.v.(type) {
	case int:
		return float64(v.v.(int))
	case string:
		r, err := strconv.ParseFloat(v.v.(string), 64)
		if err != nil {
			return 0
		}
		return r
	case float64:
		return v.v.(float64)
	default:
		return 0
	}
}

func (v *Value) String() string {
	switch v.v.(type) {
	case int:
		return strconv.Itoa(v.v.(int))
	case string:
		return v.v.(string)
	case float64:
		return strconv.FormatFloat(v.v.(float64), 'f', -1, 64)
	default:
		return ""
	}
}

func (v *Value) Bool() bool {
	switch v.v.(type) {
	case bool:
		return v.v.(bool)
	default:
		return strings.ToUpper(v.String()) == "TRUE"
	}
}

type Json struct {
	any jsoniter.Any
}

func (j *Json) Reset(data interface{}) jsoniter.Any {
	bts, err := jsoniter.Marshal(data)
	if err != nil {
		return j.any
	}
	j.any = jsoniter.Get(bts)
	return j.any
}

// GetByID 获取
func (j *Json) Iter() jsoniter.Any {
	return j.any
}

func (j *Json) Bytes() []byte {
	return j.Bytes()
}

func (j *Json) String() string {
	return j.any.ToString()
}

func (j *Json) Path(path ...interface{}) jsoniter.Any {
	return j.any.Get(path...)
}

func (j *Json) Array(path ...interface{}) Array {
	var result []jsoniter.Any
	var val = j.any.Get(path...)
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
	Response http.ResponseWriter
	Request  *http.Request
	Params   *Params
	Context  interface{}
	Query    *Store
	Form     *Store
	Json     *Json
	Files    *Files

	maxMemory int64
}

func NewStream(w http.ResponseWriter, r *http.Request, p *Params) *Stream {
	return &Stream{Response: w, Request: r, Params: p}
}

func (stream *Stream) SetMaxMemory(maxMemory int64) {
	stream.maxMemory = maxMemory
}

func (stream *Stream) SetHeader(header string, content string) {
	stream.Response.Header().Set(header, content)
}

func (stream *Stream) JsonFormat(status string, code int, msg interface{}) exception.ErrorFunc {
	return exception.New(stream.EndJson(HttpJsonResponse{status, code, msg}))
}

func (stream *Stream) End(data interface{}) error {

	var err error

	switch data.(type) {
	case []byte:
		err = stream.EndBytes(data.([]byte))
	case string:
		err = stream.EndString(data.(string))
	default:
		_, err = fmt.Fprint(stream.Response, data)
	}

	return err
}

func (stream *Stream) EndJson(data interface{}) error {
	stream.SetHeader("Content-Type", "application/json")
	return jsoniter.NewEncoder(stream.Response).Encode(data)
}

func (stream *Stream) EndString(data string) error {
	_, err := stream.Response.Write([]byte(data))
	return err
}

func (stream *Stream) EndBytes(data []byte) error {
	_, err := stream.Response.Write(data)
	return err
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

func (stream *Stream) ParseJson() *Json {

	if stream.Json != nil {
		return stream.Json
	}

	var json = new(Json)

	jsonBody, err := ioutil.ReadAll(stream.Request.Body)
	if err != nil {
		return json
	}

	json.any = jsoniter.Get(jsonBody)

	stream.Json = json

	return stream.Json
}

func (stream *Stream) ParseFiles() *Files {

	if stream.Files != nil {
		return stream.Files
	}

	var file = new(Files)

	err := stream.Request.ParseMultipartForm(stream.maxMemory)
	if err != nil {
		return file
	}

	var data = stream.Request.MultipartForm.File

	file.files = data

	stream.Files = file

	return file
}

func (stream *Stream) ParseMultipart() *Store {

	if stream.Form != nil {
		return stream.Form
	}

	var form = new(Store)

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

func (stream *Stream) ParseQuery() *Store {

	if stream.Query != nil {
		return stream.Query
	}

	var query = new(Store)

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

func (stream *Stream) ParseForm() *Store {

	if stream.Form != nil {
		return stream.Form
	}

	var form = new(Store)

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

func (stream *Stream) AutoGet(key string) *Value {
	if stream.Query != nil {
		var val = stream.Query.Get(key)
		if val.v != nil {
			return val
		}
	}
	if stream.Form != nil {
		var val = stream.Form.Get(key)
		if val.v != nil {
			return val
		}
	}
	if stream.Json != nil {
		return &Value{v: stream.Json.Path(key).GetInterface()}
	}
	return &Value{}
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

func (stream *Stream) Scheme() string {
	var scheme = "http"
	if stream.Request.TLS != nil {
		scheme = "https"
	}
	return scheme
}

type Store struct {
	keys   []string
	values []interface{}
}

func (store *Store) Has(key string) bool {
	for i := 0; i < len(store.keys); i++ {
		if store.keys[i] == key {
			return true
		}
	}
	return false
}

func (store *Store) IsEmpty(key string) bool {
	return store.Get(key).v == nil
}

func (store *Store) Get(key string) *Value {
	var val = &Value{}
	for i := 0; i < len(store.keys); i++ {
		if store.keys[i] == key {
			val.v = store.values[i]
			return val
		}
	}
	return val
}

func (store *Store) Add(key string, value interface{}) {
	store.keys = append(store.keys, key)
	store.values = append(store.values, value)
}

func (store *Store) Remove(key string) {
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

func (store *Store) Keys() []string {
	return store.keys
}

func (store *Store) Values() []interface{} {
	return store.values
}

func (store *Store) String() string {

	var buff bytes.Buffer

	for i := 0; i < len(store.keys); i++ {
		buff.WriteString(store.keys[i] + ":")
		switch store.values[i].(type) {
		case int:
			buff.WriteString(strconv.Itoa(store.values[i].(int)))
		case string:
			buff.WriteString(store.values[i].(string))
		case float64:
			buff.WriteString(strconv.FormatFloat(store.values[i].(float64), 'f', -1, 64))
		default:
			buff.WriteString(fmt.Sprintf("%v", store.values[i]))
		}
		buff.WriteString(" ")
	}

	if buff.Len() == 0 {
		return ""
	}

	var res = buff.String()

	return res[:len(res)-1]
}
