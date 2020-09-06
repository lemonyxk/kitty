/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-10-05 14:19
**/

package server

import (
	"fmt"
	"io/ioutil"
	http3 "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lemoyxk/kitty"
	"github.com/lemoyxk/kitty/http"
)

var httpServer = Server{}

var ts = httptest.NewServer(&httpServer)

type res struct {
	data string
	resp *http3.Response
}

func init() {
	httpServer.Use(func(next Middle) Middle {
		return func(stream *http.Stream) {
			stream.AutoParse()
			next(stream)
		}
	})
}

func Get(path string, params kitty.M) res {
	var p = ""
	for k, v := range params {
		p += fmt.Sprintf("%v=%v&", k, v)
	}
	if len(p) > 0 {
		p = p[:len(p)-1]
	}
	resp, err := http3.Get(ts.URL + path + "?" + p)
	if err != nil {
		panic(err)
	}
	defer func() { _ = resp.Body.Close() }()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return res{data: string(bodyBytes), resp: resp}
}

func Post(path string, params kitty.M) res {
	var p = ""
	for k, v := range params {
		p += fmt.Sprintf("%v=%v&", k, v)
	}
	if len(p) > 0 {
		p = p[:len(p)-1]
	}
	resp, err := http3.Post(ts.URL+path, "application/x-www-form-urlencoded", strings.NewReader(p))
	if err != nil {
		panic(err)
	}

	defer func() { _ = resp.Body.Close() }()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return res{data: string(body), resp: resp}
}

func AssertEqual(t *testing.T, cond bool, msg ...interface{}) {
	if !cond {
		t.Fatal(msg...)
	}
}

func Test_Method_Get(t *testing.T) {

	var httpServerRouter = &Router{}

	httpServerRouter.Route("GET", "/hello").Handler(func(stream *http.Stream) error {
		AssertEqual(t, stream.Query.First("a").String() == "1")
		return stream.EndString("hello world!")
	})

	httpServer.SetRouter(httpServerRouter)

	AssertEqual(t, Get("/hello", kitty.M{"a": 1}).data == "hello world!")
}

func Test_Method_Post(t *testing.T) {

	var httpServerRouter = &Router{}

	httpServerRouter.Route("POST", "/hello").Handler(func(stream *http.Stream) error {
		AssertEqual(t, stream.Form.First("a").String() == "2")
		return stream.End("hello group!")
	})

	httpServer.SetRouter(httpServerRouter)

	AssertEqual(t, Post("/hello", kitty.M{"a": 2}).data == "hello group!")
}

func Test_Method_NotFound(t *testing.T) {
	AssertEqual(t, Post("/not-found", nil).resp.StatusCode == http3.StatusNotFound)
}

func Test_Static_File(t *testing.T) {
	var httpServerRouter = &Router{}
	httpServerRouter.SetStaticPath("/", "../../example/server/public")
	httpServer.SetRouter(httpServerRouter)
	AssertEqual(t, len(Get("/1.png", nil).data) == 2853516)
	AssertEqual(t, Get("/test.txt", nil).data == "hello static!")
}
