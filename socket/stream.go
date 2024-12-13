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
	json "github.com/bytedance/sonic"
	"github.com/lemonyxk/kitty/errors"
	"github.com/lemonyxk/kitty/kitty"
	"github.com/lemonyxk/kitty/socket/protocol"
	"google.golang.org/protobuf/proto"
	"time"
)

func NewStream[T Packer](conn T, order uint32, messageType byte, code uint32, id uint64, route []byte, body []byte) *Stream[T] {
	return &Stream[T]{
		Time: time.Now(),
		data: body,
		sender: &sender[T]{
			conn: conn, code: code, messageID: id,
			order: order, messageType: messageType, event: string(route),
		},
	}
}

type Stream[T Packer] struct {
	*sender[T]

	data []byte

	Time time.Time

	Context kitty.Context
	Logger  kitty.Logger
	Params  Params

	//Node *router.Node[*Stream[T]]

}

func (s *Stream[T]) Data() []byte {
	return s.data
}

func (s *Stream[T]) Respond(data any) error {
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

func (s *Stream[T]) Emit(event string, data []byte) error {
	return s.conn.Pack(s.order, protocol.Bin, s.code, s.messageID, []byte(event), data)
}

func (s *Stream[T]) JsonEmit(event string, data any) error {
	msg, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return s.conn.Pack(s.order, protocol.Json, s.code, s.messageID, []byte(event), msg)
}

func (s *Stream[T]) ProtoBufEmit(event string, data proto.Message) error {
	msg, err := proto.Marshal(data)
	if err != nil {
		return err
	}
	return s.conn.Pack(s.order, protocol.ProtoBuf, s.code, s.messageID, []byte(event), msg)
}
