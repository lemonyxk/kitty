package ws

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

type Query struct {
	params map[string]string
}

type rs struct {
	Response http.ResponseWriter
	Request  *http.Request
	Context  interface{}
	Params   Params
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

func (stream *Stream) ParseQuery() (*Query, error) {

	var params = stream.Request.URL.RawQuery

	parse, err := url.ParseQuery(params)
	if err != nil {
		return nil, err
	}

	var data = make(map[string]string)

	for k, v := range parse {
		data[k] = v[0]
	}

	var query Query

	query.params = data

	return &query, nil
}

func (stream *Stream) ParseForm() (*Query, error) {

	err := stream.Request.ParseForm()
	if err != nil {
		return nil, err
	}

	var parse = stream.Request.PostForm

	var data = make(map[string]string)

	for k, v := range parse {
		data[k] = v[0]
	}

	var query Query

	query.params = data

	return &query, nil
}

func (q *Query) Get(key string) *value {

	var val = &value{}

	if v, ok := q.params[key]; ok {
		val.v = v
		return val
	}

	return val
}
