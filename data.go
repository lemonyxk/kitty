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

type HttpJsonResponse struct {
	Status string      `json:"status"`
	Code   int         `json:"code"`
	Msg    interface{} `json:"msg"`
}

func H(status string, code int, msg interface{}) HttpJsonResponse {
	return HttpJsonResponse{Status: status, Code: code, Msg: msg}
}

type SocketJsonResponse struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

func S(event string, data interface{}) SocketJsonResponse {
	return SocketJsonResponse{Event: event, Data: data}
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
