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

import "github.com/golang/protobuf/proto"

type J struct {
	Status string      `json:"status"`
	Code   int         `json:"code"`
	Msg    interface{} `json:"msg"`
}

type F struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type Receive struct {
	Context Context
	Params  *Params
	Message *ReceivePackage
}

type ReceivePackage struct {
	MessageType int
	Event       string
	Message     []byte
	ProtoType   int
}

type JsonPackage struct {
	Event   string
	Message interface{}
}

type ProtoBufPackage struct {
	Event   string
	Message proto.Message
}

type PushPackage struct {
	MessageType int
	FD          uint32
	Message     []byte
}

type M map[string]interface{}
