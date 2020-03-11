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

type JsonFormat struct {
	Status string      `json:"status"`
	Code   int         `json:"code"`
	Msg    interface{} `json:"msg"`
}

func JM(status string, code int, msg interface{}) *JsonFormat {
	return &JsonFormat{Status: status, Code: code, Msg: msg}
}

type JsonMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

func EM(event string, data interface{}) *JsonMessage {
	return &JsonMessage{Event: event, Data: data}
}

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
	Event   string
	Message interface{}
}

type ProtoBufPackage struct {
	Event   string
	Message proto.Message
}

type PushInfo struct {
	MessageType int
	FD          uint32
	Message     []byte
}

type M map[string]interface{}

type A []interface{}

type Context interface{}
