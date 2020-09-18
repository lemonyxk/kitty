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
	http3 "net/http"
	"net/http/httptest"
	"testing"

	"github.com/lemoyxk/kitty"
	"github.com/lemoyxk/kitty/http"
)

var httpServer *Server

var ts *httptest.Server

func TestMain(t *testing.M) {

	httpServer = &Server{}
	ts = httptest.NewServer(httpServer)

	httpServer.Use(func(next Middle) Middle {
		return func(stream *http.Stream) {
			stream.AutoParse()
			next(stream)
		}
	})

	t.Run()
}

func Test_Method_Get(t *testing.T) {

	var httpServerRouter = &Router{}

	httpServerRouter.Route("GET", "/hello").Handler(func(stream *http.Stream) error {
		kitty.AssertEqual(t, stream.Query.First("a").String() == "1")
		return stream.EndString("hello world!")
	})

	httpServer.SetRouter(httpServerRouter)
	kitty.AssertEqual(t, kitty.TestGet(ts.URL+"/hello", kitty.M{"a": 1}).Data == "hello world!")
}

func Test_Method_Post(t *testing.T) {

	var httpServerRouter = &Router{}

	httpServerRouter.Route("POST", "/hello").Handler(func(stream *http.Stream) error {
		kitty.AssertEqual(t, stream.Form.First("a").String() == "2")
		return stream.End("hello group!")
	})

	httpServer.SetRouter(httpServerRouter)

	kitty.AssertEqual(t, kitty.TestPost(ts.URL+"/hello", kitty.M{"a": 2}).Data == "hello group!")
}

func Test_Method_NotFound(t *testing.T) {
	kitty.AssertEqual(t, kitty.TestPost(ts.URL+"/not-found", nil).Response.StatusCode == http3.StatusNotFound)
}

func Test_Static_File(t *testing.T) {
	var httpServerRouter = &Router{}
	httpServerRouter.SetStaticPath("/", "../../example/server/public")
	httpServer.SetRouter(httpServerRouter)
	kitty.AssertEqual(t, len(kitty.TestGet(ts.URL+"/1.png", nil).Data) == 2853516)
	kitty.AssertEqual(t, kitty.TestGet(ts.URL+"/test.txt", nil).Data == "hello static!")
}
