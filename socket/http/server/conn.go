/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2021-02-11 16:22
**/

package server

type Conn interface {
	Server() *Server
}

type conn struct {
	server *Server
}

func (c *conn) Server() *Server {
	return c.server
}
