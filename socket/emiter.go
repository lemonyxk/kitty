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

	SetCode(code int)
	Code() int
	SetMessageID(messageID int64)
	MessageID() int64
	Conn() T
}
