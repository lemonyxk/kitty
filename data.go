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

type JsonMessage struct {
	Status string      `json:"status"`
	Code   int         `json:"code"`
	Msg    interface{} `json:"msg"`
}

func H(status string, code int, msg interface{}) JsonMessage {
	return JsonMessage{Status: status, Code: code, Msg: msg}
}

type EventMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

func S(event string, data interface{}) EventMessage {
	return EventMessage{Event: event, Data: data}
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
	Message *JsonMessage
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

type Context interface{}
