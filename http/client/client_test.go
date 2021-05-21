/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-10-05 14:19
**/

package client

import (
	http3 "net/http"
	"net/http/httptest"
	"testing"

	"github.com/lemoyxk/kitty"
	"github.com/lemoyxk/kitty/http"
	"github.com/lemoyxk/kitty/http/server"
	"github.com/stretchr/testify/assert"
)

var httpServer *server.Server

var ts *httptest.Server

func TestMain(t *testing.M) {

	httpServer = &server.Server{}
	ts = httptest.NewServer(httpServer)

	httpServer.Use(func(next server.Middle) server.Middle {
		return func(stream *http.Stream) {
			stream.AutoParse()
			next(stream)
		}
	})

	t.Run()
}

func Test_Method_Get(t *testing.T) {

	var httpServerRouter = &server.Router{}

	httpServerRouter.Route("GET", "/hello").Handler(func(stream *http.Stream) error {
		assert.True(t, stream.Query.First("a").String() == "1")
		return stream.EndString("hello world!")
	})

	httpServer.SetRouter(httpServerRouter)

	var res = Get(ts.URL + "/hello").Query(kitty.M{"a": 1}).Send()
	assert.True(t, res.String() == "hello world!")
}

func Test_Method_Post(t *testing.T) {

	var httpServerRouter = &server.Router{}

	httpServerRouter.Route("POST", "/hello").Handler(func(stream *http.Stream) error {
		assert.True(t, stream.Form.First("a").String() == "2")
		return stream.End("hello group!")
	})

	httpServer.SetRouter(httpServerRouter)

	var res = Post(ts.URL + "/hello").Form(kitty.M{"a": 2}).Send()

	assert.True(t, res.String() == "hello group!")
}

func Test_Method_NotFound(t *testing.T) {
	var res = Post(ts.URL + "/not-found").Form(kitty.M{"a": 2}).Send()
	assert.True(t, res.Response().StatusCode == http3.StatusNotFound)
}

func Test_Static_File(t *testing.T) {
	var httpServerRouter = &server.Router{}
	httpServerRouter.SetStaticPath("/", "../../example/server/public")
	httpServer.SetRouter(httpServerRouter)
	assert.True(t, len(Get(ts.URL+"/1.png").Query().Send().Bytes()) == 2853516)
	assert.True(t, Get(ts.URL+"/test.txt").Query().Send().String() == "hello static!")
}
