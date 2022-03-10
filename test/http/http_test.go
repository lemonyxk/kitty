/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-10-05 14:19
**/

package http

import (
	http3 "net/http"
	"net/http/httptest"
	"testing"

	"github.com/lemoyxk/kitty/v2"
	"github.com/lemoyxk/kitty/v2/http"
	"github.com/lemoyxk/kitty/v2/http/client"
	"github.com/lemoyxk/kitty/v2/http/server"
	kitty2 "github.com/lemoyxk/kitty/v2/kitty"
	"github.com/stretchr/testify/assert"
)

var httpServer *server.Server

var ts *httptest.Server

func TestMain(t *testing.M) {

	httpServer = kitty.NewHttpServer("127.0.0.1:12345")
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

	var res = client.Get(ts.URL + "/hello").Query(kitty2.M{"a": 1}).Send()
	assert.True(t, res.String() == "hello world!")
}

func Test_Method_Post(t *testing.T) {

	var httpServerRouter = &server.Router{}

	httpServerRouter.Route("POST", "/hello").Handler(func(stream *http.Stream) error {
		assert.True(t, stream.Form.First("a").String() == "2")
		return stream.End("hello group!")
	})

	httpServer.SetRouter(httpServerRouter)

	var res = client.Post(ts.URL + "/hello").Form(kitty2.M{"a": 2}).Send()

	assert.True(t, res.String() == "hello group!")
}

func Test_Method_NotFound(t *testing.T) {
	var res = client.Post(ts.URL + "/not-found").Form(kitty2.M{"a": 2}).Send()
	assert.True(t, res.Response().StatusCode == http3.StatusNotFound)
}

func Test_Static_File(t *testing.T) {
	var httpServerRouter = &server.Router{}
	httpServerRouter.SetStaticPath("/", "", http3.Dir("../../example/public"))
	httpServer.SetRouter(httpServerRouter)
	assert.True(t, len(client.Get(ts.URL+"/1.png").Query().Send().Bytes()) == 2853516)
	assert.True(t, client.Get(ts.URL+"/test.txt").Query().Send().String() == "hello static!")
}
