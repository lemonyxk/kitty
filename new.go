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
	httpClient "github.com/lemonyxk/kitty/v2/http/client"
	httpServer "github.com/lemonyxk/kitty/v2/http/server"
	tcpClient "github.com/lemonyxk/kitty/v2/socket/tcp/client"
	tcpServer "github.com/lemonyxk/kitty/v2/socket/tcp/server"
	udpClient "github.com/lemonyxk/kitty/v2/socket/udp/client"
	udpServer "github.com/lemonyxk/kitty/v2/socket/udp/server"
	webSocketClient "github.com/lemonyxk/kitty/v2/socket/websocket/client"
	webSocketServer "github.com/lemonyxk/kitty/v2/socket/websocket/server"
)

// HTTP

func NewHttpServer(addr string) *httpServer.Server {
	return &httpServer.Server{Addr: addr}
}

func NewHttpServerRouter() *httpServer.Router {
	return &httpServer.Router{}
}

func NewHttpClientProgress() *httpClient.Progress {
	return &httpClient.Progress{}
}

func NewHttpClient() *httpClient.Client {
	return &httpClient.Client{}
}

// WEB SOCKET

func NewWebSocketClientRouter() *webSocketClient.Router {
	return &webSocketClient.Router{}
}

func NewWebSocketClient(addr string) *webSocketClient.Client {
	return &webSocketClient.Client{Addr: addr}
}

func NewWebSocketServer(addr string) *webSocketServer.Server {
	return &webSocketServer.Server{Addr: addr}
}

func NewWebSocketServerRouter() *webSocketServer.Router {
	return &webSocketServer.Router{}
}

// UDP

func NewUdpServer(addr string) *udpServer.Server {
	return &udpServer.Server{Addr: addr}
}

func NewUdpServerRouter() *udpServer.Router {
	return &udpServer.Router{}
}

func NewUdpClient(addr string) *udpClient.Client {
	return &udpClient.Client{Addr: addr}
}

func NewUdpClientRouter() *udpClient.Router {
	return &udpClient.Router{}
}

// TCP

func NewTcpServer(addr string) *tcpServer.Server {
	return &tcpServer.Server{Addr: addr}
}

func NewTcpServerRouter() *tcpServer.Router {
	return &tcpServer.Router{}
}

func NewTcpClient(addr string) *tcpClient.Client {
	return &tcpClient.Client{Addr: addr}
}

func NewTcpClientRouter() *tcpClient.Router {
	return &tcpClient.Router{}
}
