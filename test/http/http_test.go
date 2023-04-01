/**
* @program: lemon
*
* @description:
*
* @author: lemon
*
* @create: 2019-10-05 14:19
**/

package http

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	http2 "net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/lemonyxk/kitty"
	"github.com/lemonyxk/kitty/example/protobuf/hello"
	kitty2 "github.com/lemonyxk/kitty/kitty"
	"github.com/lemonyxk/kitty/router"
	"github.com/lemonyxk/kitty/socket/http"
	"github.com/lemonyxk/kitty/socket/http/client"
	"github.com/lemonyxk/kitty/socket/http/server"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

var httpServer *server.Server

var ts *httptest.Server

var httpsServer *server.Server

// var tss *httptest.Server

// create by mkcert
// var certFile = "../../example/ssl/localhost+2.pem"
// var keyFile = "../../example/ssl/localhost+2-key.pem"

// can not share the ca file
// var caFile = `/Users/lemon/Library/Application Support/mkcert/rootCA.pem`

func newTls(certFile, keyFile, caFile string) *tls.Config {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}

	// Load CA cert
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	return tlsConfig
}

func TestMain(t *testing.M) {

	httpServer = kitty.NewHttpServer("127.0.0.1:12345")
	ts = httptest.NewServer(httpServer)

	httpServer.Use(func(next server.Middle) server.Middle {
		return func(stream *http.Stream) {
			stream.Parser.Auto()
			next(stream)
		}
	})

	httpsServer = kitty.NewHttpServer("127.0.0.1:12346")
	// fake server, but need ca file
	// tss = httptest.NewUnstartedServer(httpsServer)
	// tss.TLS = newTls(certFile, keyFile, caFile)
	// tss.StartTLS()

	// run the real server without ca file
	// httpsServer.CertFile = certFile
	// httpsServer.KeyFile = keyFile
	httpsServer.Use(func(next server.Middle) server.Middle {
		return func(stream *http.Stream) {
			stream.Parser.Auto()
			next(stream)
		}
	})

	go httpsServer.Start()

	var ready = make(chan bool)

	httpsServer.OnSuccess = func() {
		ready <- true
	}

	<-ready

	t.Run()
}

func Test_HTTPS_Get(t *testing.T) {

	var httpServerRouter = &router.Router[*http.Stream]{}

	httpServerRouter.Method("GET").Route("/hello").Handler(func(stream *http.Stream) error {
		var res = stream.Query.First("a").String()
		assert.True(t, res == "1", res)
		return stream.Sender.String("hello world!")
	})

	httpsServer.SetRouter(httpServerRouter)

	// assert.True(t, strings.HasPrefix(tss.URL, "https"), tss.URL)

	var res = client.Get(`http://127.0.0.1:12346` + "/hello").Query(kitty2.M{"a": 1}).Send()
	assert.True(t, res.String() == "hello world!", res.Error())
}

func Test_HTTP_Get(t *testing.T) {

	var httpServerRouter = &router.Router[*http.Stream]{}

	httpServerRouter.Method("GET").Route("/hello").Handler(func(stream *http.Stream) error {
		assert.True(t, stream.Query.First("a").String() == "1")
		return stream.Sender.String("hello world!")
	})

	httpServer.SetRouter(httpServerRouter)

	var res = client.Get(ts.URL + "/hello").Query(kitty2.M{"a": 1}).Send()
	assert.True(t, res.String() == "hello world!")
}

func Test_HTTP_Post(t *testing.T) {

	var httpServerRouter = &router.Router[*http.Stream]{}

	httpServerRouter.Method("POST").Route("/hello").Handler(func(stream *http.Stream) error {
		assert.True(t, stream.Form.First("a").String() == "2")
		return stream.Sender.String("hello group!")
	})

	httpServer.SetRouter(httpServerRouter)

	var res = client.Post(ts.URL + "/hello").Form(kitty2.M{"a": 2}).Send()

	assert.True(t, res.String() == "hello group!")
}

func Test_HTTP_PostJson(t *testing.T) {

	var httpServerRouter = &router.Router[*http.Stream]{}

	httpServerRouter.Method("POST").Route("/hello").Handler(func(stream *http.Stream) error {
		assert.True(t, stream.Json.Get("a").String() == "2")
		return stream.Sender.String("hello group!")
	})

	httpServer.SetRouter(httpServerRouter)

	var res = client.Post(ts.URL + "/hello").Json(kitty2.M{"a": 2}).Send()

	assert.True(t, res.String() == "hello group!")
}

func Test_HTTP_Head(t *testing.T) {

	var httpServerRouter = &router.Router[*http.Stream]{}

	httpServerRouter.Method("HEAD").Route("/hello").Handler(func(stream *http.Stream) error {
		assert.True(t, stream.Query.First("a").String() == "1")
		return stream.Sender.String("hello world!")
	})

	httpServer.SetRouter(httpServerRouter)

	var res = client.Head(ts.URL + "/hello").Query(kitty2.M{"a": 1}).Send()
	assert.True(t, res.String() == "", res.Response())
	assert.True(t, res.Response().ContentLength == 12)
}

func Test_HTTP_Put(t *testing.T) {

	var httpServerRouter = &router.Router[*http.Stream]{}

	httpServerRouter.Method("PUT").Route("/hello").Handler(func(stream *http.Stream) error {
		assert.True(t, stream.Form.First("a").String() == "1")
		return stream.Sender.String("hello world!")
	})

	httpServer.SetRouter(httpServerRouter)

	var res = client.Put(ts.URL + "/hello").Form(kitty2.M{"a": 1}).Send()
	assert.True(t, res.String() == "hello world!")
}

func Test_HTTP_Patch(t *testing.T) {

	var httpServerRouter = &router.Router[*http.Stream]{}

	httpServerRouter.Method("PATCH").Route("/hello").Handler(func(stream *http.Stream) error {
		assert.True(t, stream.Form.First("a").String() == "1")
		return stream.Sender.String("hello world!")
	})

	httpServer.SetRouter(httpServerRouter)

	var res = client.Patch(ts.URL + "/hello").Form(kitty2.M{"a": 1}).Send()
	assert.True(t, res.String() == "hello world!")
}

func Test_HTTP_Delete(t *testing.T) {

	var httpServerRouter = &router.Router[*http.Stream]{}

	httpServerRouter.Method("DELETE").Route("/hello").Handler(func(stream *http.Stream) error {
		assert.True(t, stream.Form.First("a").String() == "1", stream.Form.String())
		assert.True(t, stream.Form.First("b").String() == "2", stream.Form.String())
		return stream.Sender.String("hello world!")
	})

	httpServer.SetRouter(httpServerRouter)

	var res = client.Delete(ts.URL + "/hello?b=2").Form(kitty2.M{"a": 1}).Send()
	assert.True(t, res.String() == "hello world!")
}

func Test_HTTP_Options(t *testing.T) {

	var httpServerRouter = &router.Router[*http.Stream]{}

	httpServerRouter.Method("OPTIONS").Route("/hello").Handler(func(stream *http.Stream) error {
		assert.True(t, stream.Query.First("a").String() == "1", stream.Query.String())
		return stream.Sender.Respond(http2.StatusNoContent, "hello world!")
	})

	httpServer.SetRouter(httpServerRouter)

	var res = client.Options(ts.URL + "/hello").Query(kitty2.M{"a": 1}).Send()
	assert.True(t, res.Response().StatusCode == http2.StatusNoContent)
}

func Test_HTTP_Trace(t *testing.T) {

	var httpServerRouter = &router.Router[*http.Stream]{}

	httpServerRouter.Method("TRACE").Route("/hello").Handler(func(stream *http.Stream) error {
		assert.True(t, stream.Query.First("a").String() == "1")
		return stream.Sender.Respond(http2.StatusNoContent, "hello world!")
	})

	httpServer.SetRouter(httpServerRouter)

	var res = client.Trace(ts.URL + "/hello").Query(kitty2.M{"a": 1}).Send()
	assert.True(t, res.Response().StatusCode == http2.StatusNoContent)
}

func Test_HTTP_Multipart(t *testing.T) {

	var httpServerRouter = &router.Router[*http.Stream]{}

	httpServerRouter.Method("POST").Route("/PostFile").Handler(func(stream *http.Stream) error {
		assert.True(t, stream.Files.First("file").Filename == "1.png")
		assert.True(t, stream.Files.First("file").Size == 2853516)
		assert.True(t, stream.Files.First("file1") == nil)
		assert.True(t, stream.Form.First("a").Int() == 1)
		return stream.Sender.String("hello PostFile!")
	})

	httpServer.SetRouter(httpServerRouter)

	var f, err = os.Open("../../example/http/public/1.png")
	assert.True(t, err == nil, err)

	var res = client.Post(ts.URL + "/PostFile").Multipart(kitty2.M{"file": f, "a": 1}).Send()

	assert.True(t, res.String() == "hello PostFile!")
}

func Test_HTTP_Params(t *testing.T) {

	var httpServerRouter = &router.Router[*http.Stream]{}

	httpServerRouter.Method("POST").Route("/Params/:id/hello").Handler(func(stream *http.Stream) error {
		assert.True(t, stream.Params.Get("id") == stream.Form.First("a").String())
		return stream.Sender.String("hello Params!")
	})

	httpServer.SetRouter(httpServerRouter)

	var res = client.Post(ts.URL + "/Params/1/hello").Form(kitty2.M{"a": "1"}).Send()
	assert.True(t, res.String() == "hello Params!")

	res = client.Post(ts.URL + "/Params/2/hello").Form(kitty2.M{"a": "2"}).Send()
	assert.True(t, res.String() == "hello Params!")
}

func Test_HTTP_Protobuf(t *testing.T) {

	var httpServerRouter = &router.Router[*http.Stream]{}

	httpServerRouter.Method("POST").Route("/proto").Handler(func(stream *http.Stream) error {
		var res hello.AwesomeMessage
		var msg = stream.Protobuf.Bytes()
		var err = proto.Unmarshal(msg, &res)
		if err != nil {
			return stream.Sender.String(err.Error())
		}

		assert.True(t, res.AwesomeField == "1", res.String())
		assert.True(t, res.AwesomeKey == "2", res.String())

		return stream.Sender.String("hello proto!")
	})

	httpServer.SetRouter(httpServerRouter)

	var msg = hello.AwesomeMessage{
		AwesomeField: "1",
		AwesomeKey:   "2",
	}

	var res = client.Post(ts.URL + "/proto").Protobuf(&msg).Send()

	assert.True(t, res.String() == "hello proto!", res)
}

func Test_HTTP_NotFound(t *testing.T) {
	var res = client.Post(ts.URL + "/not-found").Form(kitty2.M{"a": 2}).Send()
	assert.True(t, res.Response().StatusCode == http2.StatusNotFound)
}

func Test_HTTP_Static(t *testing.T) {
	var httpServerRouter = &router.Router[*http.Stream]{}
	var httpServerStaticRouter = &server.StaticRouter{}
	httpServerStaticRouter.SetStaticPath("/", "", http2.Dir("../../example/http/public"))
	httpServer.SetRouter(httpServerRouter)
	httpServer.SetStaticRouter(httpServerStaticRouter)
	assert.True(t, len(client.Get(ts.URL+"/1.png").Query().Send().Bytes()) == 2853516)
	assert.True(t, client.Get(ts.URL+"/test.txt").Query().Send().String() == "hello static!")
}
