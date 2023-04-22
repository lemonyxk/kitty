/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2023-04-22 17:34
**/

package socket

import (
	"sync/atomic"

	jsoniter "github.com/json-iterator/go"
	"github.com/lemonyxk/kitty/socket/protocol"
	"google.golang.org/protobuf/proto"
)

func NewSender[T Packer](conn T) Emitter[T] {
	return &sender[T]{conn: conn}
}

type sender[T Packer] struct {
	conn      T
	code      int
	messageID int64
}

func (s *sender[T]) Conn() T {
	return s.conn
}

func (s *sender[T]) MessageID() int64 {
	return s.messageID
}

func (s *sender[T]) SetMessageID(messageID int64) {
	s.messageID = messageID
}

func (s *sender[T]) Code() int {
	return s.code
}

func (s *sender[T]) SetCode(code int) {
	s.code = code
}

func (s *sender[T]) Emit(event string, data []byte) error {
	return s.conn.Pack(protocol.Bin, s.code, atomic.AddInt64(&s.messageID, 1), []byte(event), data)
}

func (s *sender[T]) JsonEmit(event string, data any) error {
	msg, err := jsoniter.Marshal(data)
	if err != nil {
		return err
	}
	return s.conn.Pack(protocol.Bin, s.code, atomic.AddInt64(&s.messageID, 1), []byte(event), msg)
}

func (s *sender[T]) ProtoBufEmit(event string, data proto.Message) error {
	msg, err := proto.Marshal(data)
	if err != nil {
		return err
	}
	return s.conn.Pack(protocol.Bin, s.code, atomic.AddInt64(&s.messageID, 1), []byte(event), msg)
}
