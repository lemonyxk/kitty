/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2021-10-13 02:23
**/

package kitty

import (
	"github.com/lemonyxk/kitty/router"
	"github.com/lemonyxk/kitty/socket"
	"github.com/lemonyxk/kitty/socket/http"
	httpClient "github.com/lemonyxk/kitty/socket/http/client"
	httpServer "github.com/lemonyxk/kitty/socket/http/server"
	tcpClient "github.com/lemonyxk/kitty/socket/tcp/client"
	tcpServer "github.com/lemonyxk/kitty/socket/tcp/server"
	udpClient "github.com/lemonyxk/kitty/socket/udp/client"
	udpServer "github.com/lemonyxk/kitty/socket/udp/server"
	webSocketClient "github.com/lemonyxk/kitty/socket/websocket/client"
	webSocketServer "github.com/lemonyxk/kitty/socket/websocket/server"
)

// HTTP

func NewHttpServer(addr string) *httpServer.Server {
	return &httpServer.Server{Addr: addr}
}

func NewHttpServerRouter() *router.Router[*http.Stream] {
	return &router.Router[*http.Stream]{}
}

func NewHttpServerStaticRouter() *httpServer.StaticRouter {
	return &httpServer.StaticRouter{}
}

func NewHttpClientProgress() *httpClient.Progress {
	return &httpClient.Progress{}
}

func NewHttpClient() *httpClient.Client {
	return &httpClient.Client{}
}

// WEB SOCKET

func NewWebSocketClientRouter() *router.Router[*socket.Stream[webSocketClient.Conn]] {
	return &router.Router[*socket.Stream[webSocketClient.Conn]]{}
}

func NewWebSocketClient(addr string) *webSocketClient.Client {
	return &webSocketClient.Client{Addr: addr}
}

func NewWebSocketServer(addr string) *webSocketServer.Server {
	return &webSocketServer.Server{Addr: addr}
}

func NewWebSocketServerRouter() *router.Router[*socket.Stream[webSocketServer.Conn]] {
	return &router.Router[*socket.Stream[webSocketServer.Conn]]{}
}

// UDP

func NewUdpServer(addr string) *udpServer.Server {
	return &udpServer.Server{Addr: addr}
}

func NewUdpServerRouter() *router.Router[*socket.Stream[udpServer.Conn]] {
	return &router.Router[*socket.Stream[udpServer.Conn]]{}
}

func NewUdpClient(addr string) *udpClient.Client {
	return &udpClient.Client{Addr: addr}
}

func NewUdpClientRouter() *router.Router[*socket.Stream[udpClient.Conn]] {
	return &router.Router[*socket.Stream[udpClient.Conn]]{}
}

// TCP

func NewTcpServer(addr string) *tcpServer.Server {
	return &tcpServer.Server{Addr: addr}
}

func NewTcpServerRouter() *router.Router[*socket.Stream[tcpServer.Conn]] {
	return &router.Router[*socket.Stream[tcpServer.Conn]]{}
}

func NewTcpClient(addr string) *tcpClient.Client {
	return &tcpClient.Client{Addr: addr}
}

func NewTcpClientRouter() *router.Router[*socket.Stream[tcpClient.Conn]] {
	return &router.Router[*socket.Stream[tcpClient.Conn]]{}
}
