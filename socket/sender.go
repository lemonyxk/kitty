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
	json "github.com/lemonyxk/kitty/json"
	"sync/atomic"

	"github.com/lemonyxk/kitty/errors"
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
	conn        T
	event       string
	code        uint32
	messageID   uint64
	order       uint32
	messageType byte
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

func (s *sender[T]) MessageType() byte {
	return s.messageType
}

func (s *sender[T]) SetMessageType(messageType byte) {
	s.messageType = messageType
}

func (s *sender[T]) Event() string {
	return s.event
}

func (s *sender[T]) Emit(event string, data []byte) error {
	return s.conn.Pack(s.order, protocol.Bin, s.code, atomic.AddUint64(&s.messageID, 1), []byte(event), data)
}

func (s *sender[T]) JsonEmit(event string, data any) error {
	msg, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return s.conn.Pack(s.order, protocol.Json, s.code, atomic.AddUint64(&s.messageID, 1), []byte(event), msg)
}

func (s *sender[T]) ProtoBufEmit(event string, data proto.Message) error {
	msg, err := proto.Marshal(data)
	if err != nil {
		return err
	}
	return s.conn.Pack(s.order, protocol.ProtoBuf, s.code, atomic.AddUint64(&s.messageID, 1), []byte(event), msg)
}

func (s *sender[T]) Respond(data any) error {
	switch s.messageType {
	case protocol.Json:
		return s.JsonEmit(s.event, data)
	case protocol.ProtoBuf:
		msg, ok := data.(proto.Message)
		if !ok {
			return errors.New("data must be proto.Message")
		}
		return s.ProtoBufEmit(s.event, msg)
	case protocol.Bin:
		msg, ok := data.([]byte)
		if !ok {
			return errors.New("data must be []byte")
		}
		return s.Emit(s.event, msg)
	}
	return errors.New("unknown message type")
}
