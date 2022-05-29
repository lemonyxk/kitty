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
	"io/ioutil"
	"log"
	http2 "net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/lemonyxk/kitty/v2"
	"github.com/lemonyxk/kitty/v2/example/protobuf"
	kitty2 "github.com/lemonyxk/kitty/v2/kitty"
	"github.com/lemonyxk/kitty/v2/router"
	"github.com/lemonyxk/kitty/v2/socket/http"
	"github.com/lemonyxk/kitty/v2/socket/http/client"
	"github.com/lemonyxk/kitty/v2/socket/http/server"
	"github.com/stretchr/testify/assert"
)

var httpServer *server.Server

var ts *httptest.Server

var httpsServer *server.Server

// var tss *httptest.Server

// create by mkcert
var certFile = "../../example/ssl/localhost+2.pem"
var keyFile = "../../example/ssl/localhost+2-key.pem"

// can not share the ca file
// var caFile = `/Users/lemo/Library/Application Support/mkcert/rootCA.pem`

func newTls(certFile, keyFile, caFile string) *tls.Config {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}

	// Load CA cert
	caCert, err := ioutil.ReadFile(caFile)
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
			stream.AutoParse()
			next(stream)
		}
	})

	httpsServer = kitty.NewHttpServer("127.0.0.1:12346")
	// fake server, but need ca file
	// tss = httptest.NewUnstartedServer(httpsServer)
	// tss.TLS = newTls(certFile, keyFile, caFile)
	// tss.StartTLS()

	// run the real server without ca file
	httpsServer.CertFile = certFile
	httpsServer.KeyFile = keyFile
	httpsServer.Use(func(next server.Middle) server.Middle {
		return func(stream *http.Stream) {
			stream.AutoParse()
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

	httpServerRouter.RouteMethod("GET", "/hello").Handler(func(stream *http.Stream) error {
		var res = stream.Query.First("a").String()
		assert.True(t, res == "1", res)
		return stream.EndString("hello world!")
	})

	httpsServer.SetRouter(httpServerRouter)

	// assert.True(t, strings.HasPrefix(tss.URL, "https"), tss.URL)

	var res = client.Get(`https://127.0.0.1:12346` + "/hello").Query(kitty2.M{"a": 1}).Send()
	assert.True(t, res.String() == "hello world!", res.LastError())
}

func Test_HTTP_Get(t *testing.T) {

	var httpServerRouter = &router.Router[*http.Stream]{}

	httpServerRouter.RouteMethod("GET", "/hello").Handler(func(stream *http.Stream) error {
		assert.True(t, stream.Query.First("a").String() == "1")
		return stream.EndString("hello world!")
	})

	httpServer.SetRouter(httpServerRouter)

	var res = client.Get(ts.URL + "/hello").Query(kitty2.M{"a": 1}).Send()
	assert.True(t, res.String() == "hello world!")
}

func Test_HTTP_Post(t *testing.T) {

	var httpServerRouter = &router.Router[*http.Stream]{}

	httpServerRouter.RouteMethod("POST", "/hello").Handler(func(stream *http.Stream) error {
		assert.True(t, stream.Form.First("a").String() == "2")
		return stream.End("hello group!")
	})

	httpServer.SetRouter(httpServerRouter)

	var res = client.Post(ts.URL + "/hello").Form(kitty2.M{"a": 2}).Send()

	assert.True(t, res.String() == "hello group!")
}

func Test_HTTP_Protobuf(t *testing.T) {

	var httpServerRouter = &router.Router[*http.Stream]{}

	httpServerRouter.RouteMethod("POST", "/proto").Handler(func(stream *http.Stream) error {
		var res awesomepackage.AwesomeMessage
		var msg = stream.Protobuf.Bytes()
		var err = proto.Unmarshal(msg, &res)
		if err != nil {
			return stream.EndString(err.Error())
		}

		assert.True(t, res.AwesomeField == "1", res)
		assert.True(t, res.AwesomeKey == "2", res)

		return stream.EndString("hello proto!")
	})

	httpServer.SetRouter(httpServerRouter)

	var msg = awesomepackage.AwesomeMessage{
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
