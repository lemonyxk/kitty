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
)

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
			for {
				buffer := make([]byte, 1)

				dl, err := conn.Read(buffer)

				// close normal
				if err == io.EOF {
					break
				}

				// close not normal
				if err != nil {
					log.Println("read error")
					return
				}

				log.Println(dl, string(buffer))
			}
		}()
	}
}
