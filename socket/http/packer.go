/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2023-06-01 15:04
**/

package http

type Packer interface {

}

type sender[T Packer] struct {
	conn        T
}

func (s *sender[T]) Conn() T {
	return s.conn
}