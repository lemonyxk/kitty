/**
* @program: lemon
*
* @description:
*
* @author: lemon
*
* @create: 2019-11-19 20:56
**/

package socket

import (
	"github.com/golang/protobuf/proto"
	"github.com/lemonyxk/kitty/v2/kitty"
)

// 0 version
// 1 message type
// 2 proto type
// 3 route len
// 4 body len
// 5 body len
// 6 body len
// 7 body len

const (
	// Version byte = 'V'

	Unknown byte = 0
	Text    byte = 1
	Bin     byte = 2
	Open    byte = 3
	Close   byte = 4
	Ping    byte = 9
	Pong    byte = 10

	// Text     int = 1
	// Json     int = 2
	// ProtoBuf int = 3
)

type Stream[T any] struct {
	Pack

	Conn T

	Context kitty.Context
	Params  kitty.Params
	Logger  kitty.Logger
}

type Pack struct {
	Event string
	Data  []byte
}

type JsonPack struct {
	Event string
	Data  any
}

type ProtoBufPack struct {
	Event string
	Data  proto.Message
}
