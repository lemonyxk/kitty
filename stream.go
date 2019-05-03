package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

type Wson map[string]interface{}

type Query struct {
	params map[string]string
}

type reqs struct {
	Response http.ResponseWriter
	Request  *http.Request
	Context  interface{}
}

type Stream struct {
	reqs
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

func (q *Query) Get(key string) string {

	if v, ok := q.params[key]; ok {
		return v
	}
	return ""
}
