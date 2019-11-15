package lemo

import (
	"bytes"
	"fmt"
	"github.com/json-iterator/go"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Query struct {
	params map[string]interface{}
}

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

type Stream struct {
	Response http.ResponseWriter
	Request  *http.Request
	Context  interface{}
	Params   *Params
	Query    *Query
	Files    *Files

	url *URL
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

func (stream *Stream) SetHeader(header string, content string) {
	stream.Response.Header().Set(header, content)
}

func (stream *Stream) JsonFormat(status string, code int, msg interface{}) error {
	return stream.Json(M{"status": status, "code": code, "msg": msg})
}

func (stream *Stream) Json(data interface{}) error {

	stream.SetHeader("Content-Type", "application/json")

	return jsoniter.NewEncoder(stream.Response).Encode(data)
}

func (stream *Stream) End(data interface{}) error {

	var err error

	switch data.(type) {
	case []byte:
		_, err = stream.Response.Write(data.([]byte))
	case string:
		_, err = stream.Response.Write([]byte(data.(string)))
	default:
		_, err = fmt.Fprint(stream.Response, data)
	}

	return err
}

func (stream *Stream) ClientIP() string {

	remoteAddr := stream.Request.RemoteAddr

	if ip := stream.Request.Header.Get(XRealIP); ip != "" {
		remoteAddr = ip
	} else if ip = stream.Request.Header.Get(XForwardedFor); ip != "" {
		remoteAddr = ip
	} else {
		remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
	}

	if remoteAddr == "::1" {
		remoteAddr = "127.0.0.1"
	}

	return remoteAddr
}

func (stream *Stream) ParseJson() *Query {

	jsonBody, err := ioutil.ReadAll(stream.Request.Body)
	if err != nil {
		return nil
	}

	var data = make(map[string]interface{})

	err = jsoniter.Unmarshal(jsonBody, &data)
	if err != nil {
		return nil
	}

	var query = new(Query)

	query.params = data

	return query
}

func (stream *Stream) ParseFiles() *Files {

	if stream.Files != nil {
		return stream.Files
	}

	err := stream.Request.ParseMultipartForm(20 * 1024 * 1024)
	if err != nil {
		return nil
	}

	var data = stream.Request.MultipartForm.File

	var file = new(Files)

	file.files = data

	stream.Files = file

	return file
}

func (stream *Stream) ParseMultipart() *Query {

	if stream.Query != nil {
		return stream.Query
	}

	err := stream.Request.ParseMultipartForm(2 * 1024 * 1024)
	if err != nil {
		return nil
	}

	var parse = stream.Request.MultipartForm.Value

	var data = make(map[string]interface{})

	for k, v := range parse {
		data[k] = v[0]
	}

	var query = new(Query)

	query.params = data

	return query
}

func (stream *Stream) ParseQuery() *Query {

	if stream.Query != nil {
		return stream.Query
	}

	var params = stream.Request.URL.RawQuery

	parse, err := url.ParseQuery(params)
	if err != nil {
		return nil
	}

	var data = make(map[string]interface{})

	for k, v := range parse {
		data[k] = v[0]
	}

	var query = new(Query)

	query.params = data

	return query
}

func (stream *Stream) ParseForm() *Query {

	if stream.Query != nil {
		return stream.Query
	}

	err := stream.Request.ParseForm()
	if err != nil {
		return nil
	}

	var parse = stream.Request.PostForm

	var data = make(map[string]interface{})

	for k, v := range parse {
		data[k] = v[0]
	}

	var query = new(Query)

	query.params = data

	return query
}

func (stream *Stream) AutoParse() *Query {

	if stream.Query != nil {
		return stream.Query
	}

	var header = stream.Request.Header.Get("Content-Type")

	var query *Query

	if strings.ToUpper(stream.Request.Method) == "GET" {
		query = stream.ParseQuery()
	} else {
		if strings.HasPrefix(header, "multipart/form-data") {
			query = stream.ParseMultipart()
		} else if strings.HasPrefix(header, "application/x-www-form-urlencoded") {
			query = stream.ParseForm()
		} else if strings.HasPrefix(header, "application/json") {
			query = stream.ParseJson()
		}
	}

	if query == nil {
		query = new(Query)
		query.params = make(map[string]interface{})
	}

	stream.Query = query

	return query
}

type URL struct {
	Url         string
	Scheme      string
	Host        string
	Path        string
	QueryString string
	Fragment    string
}

func (stream *Stream) Url() *URL {

	if stream.url != nil {
		return stream.url
	}

	var buff bytes.Buffer

	var scheme = "http"

	if stream.Request.TLS != nil {
		scheme = "https"
	}

	buff.WriteString(scheme)
	buff.WriteString("://")
	buff.WriteString(stream.Request.Host)
	buff.WriteString(stream.Request.URL.Path)
	buff.WriteString(stream.Request.URL.RawQuery)
	if stream.Request.URL.Fragment != "" {
		buff.WriteString("#")
		buff.WriteString(stream.Request.URL.Fragment)
	}

	stream.url = &URL{}

	stream.url.Url = buff.String()
	stream.url.Scheme = scheme
	stream.url.Host = stream.Request.Host
	stream.url.Path = stream.Request.URL.Path
	stream.url.QueryString = stream.Request.URL.RawQuery
	stream.url.Fragment = stream.Request.URL.Fragment

	return stream.url
}

func (q *Query) Has(key string) bool {
	_, ok := q.params[key]
	return ok
}

func (q *Query) IsStringEmpty(key string) bool {
	return q.Get(key).String() == ""
}

func (q *Query) IsIntEmpty(key string) bool {
	return q.Get(key).Int() == 0
}

func (q *Query) IsFloatEmpty(key string) bool {
	return q.Get(key).Float64() == 0
}

func (q *Query) Get(key string) *Value {

	var val = &Value{}

	if v, ok := q.params[key]; ok {
		val.v = v
		return val
	}

	return val
}

func (q *Query) All() map[string]interface{} {
	return q.params
}

func (q *Query) String() string {

	var buff bytes.Buffer

	for key, value := range q.params {

		buff.WriteString(key)
		buff.WriteString(":")

		switch value.(type) {
		case int:
			buff.WriteString(strconv.Itoa(value.(int)))
		case string:
			buff.WriteString(value.(string))
		case float64:
			buff.WriteString(strconv.FormatFloat(value.(float64), 'f', -1, 64))
		default:
			buff.WriteString(fmt.Sprintf("%v", value))
		}

		buff.WriteString(" ")
	}

	if buff.Len() == 0 {
		return ""
	}

	var res = buff.String()

	return res[:len(res)-1]
}
