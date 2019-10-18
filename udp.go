/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-10-17 20:09
**/

package lemo

import (
	"log"
	"net"
)

func Udp() {

	udp, err := net.ListenPacket("udp", "127.0.0.1:5000")
	if err != nil {
		panic(err)
	}

	// err = udp.SetReadDeadline(time.Now())
	// if err != nil {
	// 	panic(err)
	// }

	var buf = make([]byte, 8096)

	for {

		n, addr, err := udp.ReadFrom(buf)
		if err != nil {
			log.Println(err)
			continue
		}

		log.Println(n, addr.String(), string(buf[0:n]))
	}

}
