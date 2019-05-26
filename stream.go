package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

type Wson map[string]interface{}

type Query struct {
	params map[string]string
}

type rs struct {
	Response http.ResponseWriter
	Request  *http.Request
	Context  interface{}
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

func (stream *Stream) Json(data interface{}) {

	stream.Response.Header().Add("Content-Type", "application/json")

	j, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		return
	}

	stream.Response.Write(j)
}

func (stream *Stream) End(data interface{}) {

	stream.Response.Header().Add("Content-Type", "text/html")

	fmt.Fprint(stream.Response, data)
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

func (q *Query) Get(key string) *value {

	var val = &value{}

	if v, ok := q.params[key]; ok {
		val.v = v
		return val
	}

	return val
}
