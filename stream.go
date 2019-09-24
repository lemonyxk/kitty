package ws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Query struct {
	params map[string]string
}

type Files struct {
	files map[string][]*multipart.FileHeader
}

type rs struct {
	Response http.ResponseWriter
	Request  *http.Request
	Context  interface{}
	Params   *Params
	URL      *URL
}

type Params struct {
	Keys   []string
	Values []string
}

func (ps *Params) ByName(name string) string {
	for i := range ps.Keys {
		if ps.Keys[i] == name {
			return ps.Values[i]
		}
	}
	return ""
}

type Stream struct {
	rs
}

type value struct {
	v string
}

func (v *value) Int() int {
	r, err := strconv.Atoi(v.v)
	if err != nil {
		return 0
	}

	return r
}

func (v *value) String() string {
	return v.v
}

func (stream *Stream) Json(data interface{}) error {

	stream.Response.Header().Add("Content-Type", "application/json")

	j, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = stream.Response.Write(j)

	return err
}

func (stream *Stream) End(data ...interface{}) error {

	stream.Response.Header().Add("Content-Type", "text/html")

	_, err := fmt.Fprint(stream.Response, data...)

	return err
}

func (stream *Stream) IP() string {

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

	var data = make(map[string]string)

	err = json.Unmarshal(jsonBody, &data)
	if err != nil {
		return nil
	}

	var query Query

	query.params = data

	return &query
}

func (stream *Stream) Files() *Files {

	err := stream.Request.ParseMultipartForm(20 * 1024 * 1024)
	if err != nil {
		return nil
	}

	var data = stream.Request.MultipartForm.File

	var query Files

	query.files = data

	return &query
}

func (stream *Stream) ParseMultipart() *Query {

	err := stream.Request.ParseMultipartForm(2 * 1024 * 1024)
	if err != nil {
		return nil
	}

	var parse = stream.Request.MultipartForm.Value

	var data = make(map[string]string)

	for k, v := range parse {
		data[k] = v[0]
	}

	var query Query

	query.params = data

	return &query
}

func (stream *Stream) ParseQuery() *Query {

	var params = stream.Request.URL.RawQuery

	parse, err := url.ParseQuery(params)
	if err != nil {
		return nil
	}

	var data = make(map[string]string)

	for k, v := range parse {
		data[k] = v[0]
	}

	var query Query

	query.params = data

	return &query
}

func (stream *Stream) ParseForm() *Query {

	err := stream.Request.ParseForm()
	if err != nil {
		return nil
	}

	var parse = stream.Request.PostForm

	var data = make(map[string]string)

	for k, v := range parse {
		data[k] = v[0]
	}

	var query Query

	query.params = data

	return &query
}

func (stream *Stream) AutoParse() *Query {

	if strings.ToUpper(stream.Request.Method) == "GET" {
		return stream.ParseQuery()
	}

	var header = stream.Request.Header.Get("Content-Type")

	if strings.HasPrefix(header, "multipart/form-data") {
		return stream.ParseMultipart()
	}

	if strings.HasPrefix(header, "application/x-www-form-urlencoded") {
		return stream.ParseForm()
	}

	if strings.HasPrefix(header, "application/json") {
		return stream.ParseJson()
	}

	return nil
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

	if stream.URL != nil {
		return stream.URL
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

	stream.URL = &URL{}

	stream.URL.Url = buff.String()
	stream.URL.Scheme = scheme
	stream.URL.Host = stream.Request.Host
	stream.URL.Path = stream.Request.URL.Path
	stream.URL.QueryString = stream.Request.URL.RawQuery
	stream.URL.Fragment = stream.Request.URL.Fragment

	return stream.URL
}

func (q *Query) Get(key string) *value {

	var val = &value{}

	if v, ok := q.params[key]; ok {
		val.v = v
		return val
	}

	return val
}

func (q *Query) String() string {

	var buff bytes.Buffer

	for key, value := range q.params {

		buff.WriteString(key)
		buff.WriteString(":")
		buff.WriteString(value)
		buff.WriteString(" ")
	}

	if buff.Len() == 0 {
		return ""
	}

	var res = buff.String()

	return res[:len(res)-1]
}
