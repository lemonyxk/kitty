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

func NewStream[T Packer](conn T, code int, messageID int64, event string, data []byte) *Stream[T] {
	return &Stream[T]{
		sender: &sender[T]{conn: conn, code: code, messageID: messageID},
		Event:  event, Data: data,
	}
}

type Stream[T Packer] struct {
	Data  []byte
	Event string

	Context kitty.Context
	Params  Params
	Logger  kitty.Logger

	*sender[T]
}

func (s *Stream[T]) Emit(event string, data []byte) error {
	return s.conn.Pack(protocol.Bin, s.code, s.messageID, []byte(event), data)
}

func (s *Stream[T]) JsonEmit(event string, data any) error {
	msg, err := jsoniter.Marshal(data)
	if err != nil {
		return err
	}
	return s.conn.Pack(protocol.Bin, s.code, s.messageID, []byte(event), msg)
}

func (s *Stream[T]) ProtoBufEmit(event string, data proto.Message) error {
	msg, err := proto.Marshal(data)
	if err != nil {
		return err
	}
	return s.conn.Pack(protocol.Bin, s.code, s.messageID, []byte(event), msg)
}
