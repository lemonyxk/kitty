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
