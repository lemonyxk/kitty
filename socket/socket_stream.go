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
	// Version byte = 'V'

	// message type
	Unknown byte = 0
	// TextData int = 1
	BinData   byte = 2
	OpenData  byte = 3
	CloseData byte = 4
	PingData  byte = 9
	PongData  byte = 10

	// proto type
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

	Context kitty.Context
	Params  kitty.Params
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
