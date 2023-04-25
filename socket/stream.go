/**
* @program: lemon
*
* @description:
*
* @author: lemon
*
* @create: 2019-11-19 20:56
**/

package socket

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/lemonyxk/kitty/kitty"
	"github.com/lemonyxk/kitty/socket/protocol"
	"google.golang.org/protobuf/proto"
)

func NewStream[T Packer](conn T, code uint32, messageID uint64, event string, data []byte) *Stream[T] {
	return &Stream[T]{
		sender: &sender[T]{conn: conn, code: code, messageID: messageID},
		event:  event, data: data,
	}
}

type Stream[T Packer] struct {
	Context kitty.Context
	Logger  kitty.Logger
	Params  Params

	data  []byte
	event string

	*sender[T]
}

func (s *Stream[T]) Data() []byte {
	return s.data
}

func (s *Stream[T]) Event() string {
	return s.event
}

func (s *Stream[T]) Emit(event string, data []byte) error {
	return s.conn.Pack(protocol.Async, protocol.Bin, s.code, s.messageID, []byte(event), data)
}

func (s *Stream[T]) JsonEmit(event string, data any) error {
	msg, err := jsoniter.Marshal(data)
	if err != nil {
		return err
	}
	return s.conn.Pack(protocol.Async, protocol.Bin, s.code, s.messageID, []byte(event), msg)
}

func (s *Stream[T]) ProtoBufEmit(event string, data proto.Message) error {
	msg, err := proto.Marshal(data)
	if err != nil {
		return err
	}
	return s.conn.Pack(protocol.Async, protocol.Bin, s.code, s.messageID, []byte(event), msg)
}
