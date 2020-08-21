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

	"github.com/lemoyxk/kitty"
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
	// Version
	Version byte = 'V'

	// message type
	Unknown  int = 0
	TextData int = 1
	BinData  int = 2
	PingData int = 9
	PongData int = 10

	// proto type
	Text     int = 1
	Json     int = 2
	ProtoBuf int = 3
)

type Stream struct {
	MessageType int
	ProtoType   int
	Event       string
	Message     []byte
	Raw         []byte

	Context kitty.Context
	Params  kitty.Params
}

type JsonPackage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type ProtoBufPackage struct {
	Event string
	Data  proto.Message
}

type PushPackage struct {
	Type int
	FD   uint32
	Data []byte
}
