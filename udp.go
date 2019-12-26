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
	"net"

	"github.com/Lemo-yxk/lemo/console"
)

func Udp() {

	var addr, err = net.ResolveUDPAddr("udp", "127.0.0.1:5000")
	if err != nil {
		panic(err)
	}

	udp, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}

	// err = udp.SetReadDeadline(time.Now())
	// if err != nil {
	// 	panic(err)
	// }

	var buf = make([]byte, 8096)

	for {

		n, addr, err := udp.ReadFromUDP(buf)
		if err != nil {
			console.Error(err)
			continue
		}

		console.Println(n, addr.String())
	}

}
