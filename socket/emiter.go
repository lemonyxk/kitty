/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2023-04-22 19:07
**/

package socket

import "google.golang.org/protobuf/proto"

type Emitter[T Packer] interface {
	JsonEmit(event string, data any) error
	ProtoBufEmit(event string, data proto.Message) error
	Emit(event string, data []byte) error

	SetCode(code uint32)
	Code() uint32
	SetMessageID(messageID uint64)
	MessageID() uint64
	SetMessageType(messageType byte)
	MessageType() byte
	Event() string
	Conn() T
}
