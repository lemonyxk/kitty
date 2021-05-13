/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2021-02-12 14:25
**/

package main

import (
	"log"
	"strings"
	"sync/atomic"
	"time"

	"github.com/lemoyxk/kitty/socket"
	client2 "github.com/lemoyxk/kitty/socket/udp/client"
	"github.com/lemoyxk/kitty/socket/udp/server"
)

func main() {

	log.SetFlags(log.Lshortfile | log.Ldate)

	var udpServer = &server.Server{Addr: "127.0.0.1:5000", HeartBeatTimeout: time.Second * 30}

	udpServer.OnMessage = func(conn *server.Conn, msg []byte) {
		// log.Println("ONMESSAGE:", string(msg))
	}

	udpServer.OnOpen = func(conn *server.Conn) {

	}

	udpServer.OnError = func(err error) {
		log.Println(err)
	}

	var udpServerRouter = &server.Router{IgnoreCase: true}

	var res uint32 = 0

	udpServerRouter.Group("/hello").Handler(func(handler *server.RouteHandler) {
		handler.Route("/world").Handler(func(conn *server.Conn, stream *socket.Stream) error {
			// log.Println("SERVER", "ID:", stream.ID, "EVENT:", stream.Event, "DATA:", string(stream.Data))
			atomic.AddUint32(&res, 1)
			return conn.Emit(socket.Pack{
				Event: stream.Event,
				Data:  []byte("i am server"),
				ID:    stream.ID,
			})
		})
	})

	udpServer.OnSuccess = func() {
		log.Println(udpServer.LocalAddr())
		go client()
	}

	go udpServer.SetRouter(udpServerRouter).Start()

	go func() {
		for {
			time.Sleep(time.Second)
			log.Println(res)
		}
	}()

	select {}
}

func client() {

	// dst := make([]byte, hex.EncodedLen(len(bts)))
	// n := hex.Encode(dst, bts)
	// log.Println(string(dst[:n]))

	// var res int32
	//
	// for i := 0; i < 50; i++ {
	// 	go func(i int) {
	// 		ip := net.ParseIP("127.0.0.1")
	// 		srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	// 		dstAddr := &net.UDPAddr{IP: ip, Port: 5001}
	// 		conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	// 		if err != nil {
	// 			fmt.Println(err)
	// 		}
	// 		defer conn.Close()
	// 		var openMsg = make([]byte, 12)
	// 		openMsg[1] = 0x1
	// 		_, err = conn.Write(openMsg)
	// 		if err != nil {
	// 			log.Println(err)
	// 		}
	//
	// 		for j := 0; j < 2000; j++ {
	// 			var bts = protocol.Encode(socket.Bin, uint32(i*10+j+1), []byte("/hello/world"), []byte("xixi"))
	// 			_, err = conn.Write(bts)
	// 			if err != nil {
	// 				log.Println(err)
	// 			}
	// 			data := make([]byte, 1024)
	// 			_, err = conn.Read(data)
	// 			if err != nil {
	// 				log.Println(err)
	// 			}
	// 			atomic.AddInt32(&res, 1)
	// 		}
	// 	}(i)
	// }
	//
	// for {
	//
	// 	log.Println(res)
	//
	// 	if res == 100000 {
	// 		log.Println("done")
	// 		break
	// 	}
	//
	// 	time.Sleep(time.Second)
	// }

	var res uint32 = 0

	for i := 0; i < 1; i++ {

		var index = i
		var client = &client2.Client{
			Addr:        "127.0.0.1:5000",
			DailTimeout: 50 * time.Second,
			// Reconnect: true, AutoHeartBeat: true, HeartBeatInterval: time.Second,
		}

		client.OnOpen = func(client *client2.Client) {
			log.Println("client open", index)
		}

		client.OnError = func(err error) {
			log.Println(err)
		}

		client.OnClose = func(client *client2.Client) {
			log.Println("client close")
		}

		client.OnSuccess = func() {
			go func() {

				for i := 0; i < 10000; i++ {
					err := client.Emit(socket.Pack{
						Event: "/hello/world",
						Data:  []byte(strings.Repeat("i", 123)),
						ID:    int64(i),
					})
					if err != nil {
						log.Println(err)
					}
				}

			}()
		}

		var router = &client2.Router{IgnoreCase: true}

		router.Group("/hello").Handler(func(handler *client2.RouteHandler) {
			handler.Route("/world").Handler(func(client *client2.Client, stream *socket.Stream) error {
				atomic.AddUint32(&res, 1)
				// log.Println("CLIENT", "ID:", stream.ID, "EVENT:", stream.Event, "DATA:", string(stream.Data))
				return nil
			})
		})

		go client.SetRouter(router).Connect()
	}

	go func() {
		for {
			time.Sleep(time.Second)
			log.Println(res)
		}
	}()
}
