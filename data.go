/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-11-19 20:56
**/

package lemo

import (
	"github.com/golang/protobuf/proto"
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

const (
	XForwardedFor = "X-Forwarded-For"
	XRealIP       = "X-Real-IP"
	Host          = "Host"
)

type Receive struct {
	Context Context
	Params  Params
	Body    *ReceivePackage
}

type ReceivePackage struct {
	MessageType int
	Event       string
	Message     []byte
	ProtoType   int
	Raw         []byte
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

type M map[string]interface{}

type A []interface{}

type Context interface{}

type Params struct {
	Keys   []string
	Values []string
}

func (ps Params) ByName(name string) string {
	for i := 0; i < len(ps.Keys); i++ {
		if ps.Keys[i] == name {
			return ps.Values[i]
		}
	}
	return ""
}
