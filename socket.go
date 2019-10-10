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
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func SocketClient() {
	conn, err := net.Dial("tcp", "127.0.0.1:5000")
	if err != nil {
		panic(err)
	}

	defer func() { _ = conn.Close() }()

	go func() {
		for i := 0; i < 100; i++ {

			// log.Println(len(msg), msg)

			_, _ = conn.Write(Pack([]byte("/hello"), []byte("world"), 1, 2))
		}
	}()

	go func() {
		for {

			buffer := make([]byte, 1024)

			n, err := conn.Read(buffer)

			// close normal
			if err == io.EOF {
				break
			}

			// close not normal
			if err != nil {
				log.Println("read error")
				return
			}

			log.Println(n, string(buffer))
		}
	}()

	// 创建信号
	signalChan := make(chan os.Signal, 1)
	// 通知
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	// 阻塞
	<-signalChan
}

func Socket() {

	netListen, err := net.Listen("tcp", ":5000")
	if err != nil {
		panic(err)
	}

	defer func() { _ = netListen.Close() }()

	for {

		conn, err := netListen.Accept()
		if err != nil {
			continue
		}

		go func() {

			var data []byte

			for {

				buffer := make([]byte, 1024)

				n, err := conn.Read(buffer)

				// close normal
				if err == io.EOF {
					log.Println("normal close")
					break
				}

				// close not normal
				if err != nil {
					log.Println("read error")
					return
				}

				data = append(data, buffer[0:n]...)

				log.Println(n, string(buffer))

				_, _ = conn.Write(buffer)
			}
		}()
	}
}
