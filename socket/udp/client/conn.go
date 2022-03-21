/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2021-10-29 18:44
**/

package client

import (
	"net"
	"sync"
	"time"

	"github.com/lemonyxk/kitty/v2/socket/udp"
)

type Conn struct {
	Conn     *net.UDPConn
	Addr     *net.UDPAddr
	LastPong time.Time
	mux      sync.RWMutex
}

func (c *Conn) LocalAddr() net.Addr {
	return c.Conn.LocalAddr()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.Addr
}

func (c *Conn) Write(message []byte) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	_, err := c.Conn.WriteToUDP(message, c.Addr)
	return err
}

func (c *Conn) WriteToAddr(message []byte, addr *net.UDPAddr) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	_, err := c.Conn.WriteToUDP(message, addr)
	return err
}

func (c *Conn) Close() error {
	_ = c.Write(udp.CloseMessage)
	return c.Conn.Close()
}

func (c *Conn) Read(b []byte) (n int, addr *net.UDPAddr, err error) {
	return c.Conn.ReadFromUDP(b)
}
