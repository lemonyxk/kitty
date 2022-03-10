/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2021-10-29 18:57
**/

package client

import (
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lemoyxk/kitty/v2/socket"
)

type Conn struct {
	Conn     *websocket.Conn
	LastPong time.Time
	mux      sync.RWMutex
}

func (c *Conn) LocalAddr() net.Addr {
	return c.Conn.LocalAddr()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *Conn) Write(message []byte) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.Conn.WriteMessage(int(socket.Bin), message)
}

func (c *Conn) Close() error {
	return c.Conn.Close()
}

func (c *Conn) Read() (int, []byte, error) {
	return c.Conn.ReadMessage()
}
