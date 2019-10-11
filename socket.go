/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-10-09 14:06
**/

package lemo

import (
	"log"
	"net"
	"sync"

	"github.com/Lemo-yxk/tire"
)

type Socket struct {
	Fd     uint32
	Conn   *net.Conn
	socket *SocketServer
}

type SocketServer struct {
	Fd          uint32
	count       uint32
	connections sync.Map
	OnClose     func(fd uint32)
	OnMessage   func(conn *Socket, messageType int, msg []byte)
	OnOpen      func(conn *Socket)
	OnError     func(err func() *Error)

	HeartBeatTimeout  int
	HeartBeatInterval int
	HandshakeTimeout  int
	ReadBufferSize    int
	WriteBufferSize   int
	WaitQueueSize     int

	Router *tire.Tire

	IgnoreCase bool

	// 连接
	connOpen chan *Socket

	// 关闭
	connClose chan *Socket

	// 写入
	connPush chan *PushPackage

	// 返回
	connBack chan error
}

func SocketTest() {

	netListen, err := net.Listen("tcp", ":5000")
	if err != nil {
		panic(err)
	}

	defer func() { _ = netListen.Close() }()

	for {

		conn, err := netListen.Accept()
		if err != nil {
			OnError(err)
			continue
		}

		go func() {

			OnOpen(conn)

			var singleMessageLen = 0

			var message []byte

			for {

				buffer := make([]byte, 1024)

				n, err := conn.Read(buffer)

				// // close normal
				// if err == io.EOF {
				// 	OnClose()
				// 	return
				// }

				// close error
				if err != nil {
					_ = conn.Close()
					OnClose()
					return
				}

				message = append(message, buffer[0:n]...)

				// read continue
				if len(message) < 8 {
					continue
				}

				for {

					// jump out and read continue
					if len(message) < 8 {
						break
					}

					// just begin
					if singleMessageLen == 0 {

						// proto error
						if !IsHeaderInvalid(message) {
							_ = conn.Close()
							OnClose()
							return
						}

						singleMessageLen = GetLen(message)
					}

					// jump out and read continue
					if len(message) < singleMessageLen {
						break
					}

					// a complete message
					OnMessage(message[0:singleMessageLen])

					// delete this message
					message = message[singleMessageLen:]

					// reset len
					singleMessageLen = 0

				}

			}

		}()
	}
}

func OnMessage(message []byte) {

	version, messageType, protoType, route, body := UnPack(message)
	log.Println(version, messageType, protoType, route, body)

}

func OnClose() {
	log.Println("close")
}

func OnOpen(conn net.Conn) {
	log.Println("open")
}

func OnError(err error) {
	log.Println("error", err)
}
