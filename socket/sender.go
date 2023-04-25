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

// NewSender returns a new Sender.
// Emit JsonEmit ProtoBufEmit it safe for concurrent use by multiple goroutines,
// but set code and messageID is not safe.
func NewSender[T Packer](conn T) Emitter[T] {
	return &sender[T]{conn: conn}
}

type sender[T Packer] struct {
	conn      T
	code      uint32
	messageID uint64
}

func (s *sender[T]) Conn() T {
	return s.conn
}

func (s *sender[T]) MessageID() uint64 {
	return s.messageID
}

func (s *sender[T]) SetMessageID(messageID uint64) {
	s.messageID = messageID
}

func (s *sender[T]) Code() uint32 {
	return s.code
}

func (s *sender[T]) SetCode(code uint32) {
	s.code = code
}

func (s *sender[T]) Emit(event string, data []byte) error {
	return s.conn.Pack(protocol.Async, protocol.Bin, s.code, atomic.AddUint64(&s.messageID, 1), []byte(event), data)
}

func (s *sender[T]) JsonEmit(event string, data any) error {
	msg, err := jsoniter.Marshal(data)
	if err != nil {
		return err
	}
	return s.conn.Pack(protocol.Async, protocol.Bin, s.code, atomic.AddUint64(&s.messageID, 1), []byte(event), msg)
}

func (s *sender[T]) ProtoBufEmit(event string, data proto.Message) error {
	msg, err := proto.Marshal(data)
	if err != nil {
		return err
	}
	return s.conn.Pack(protocol.Async, protocol.Bin, s.code, atomic.AddUint64(&s.messageID, 1), []byte(event), msg)
}
