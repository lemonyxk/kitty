/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2021-10-29 18:38
**/

package client

import (
	"net"
	"sync"
	"time"
)

type Conn struct {
	Conn     net.Conn
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
	_, err := c.Conn.Write(message)
	return err
}

func (c *Conn) Close() error {
	return c.Conn.Close()
}

func (c *Conn) Read(b []byte) (n int, err error) {
	return c.Conn.Read(b)
}
