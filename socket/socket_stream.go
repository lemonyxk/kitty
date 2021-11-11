/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-11-19 20:56
**/

package socket

import (
	"github.com/golang/protobuf/proto"
	kitty2 "github.com/lemoyxk/kitty/kitty"
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

type Pack struct {
	Event string
	Data  []byte
	ID    int64
}

type Stream struct {
	Pack

	Context kitty2.Context
	Params  kitty2.Params
	Logger  kitty2.Logger
}

type JsonPack struct {
	Event string
	Data  interface{}
	ID    int64
}

type ProtoBufPack struct {
	Event string
	Data  proto.Message
	ID    int64
}
