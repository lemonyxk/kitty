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
	"github.com/lemonyxk/kitty/errors"
	json "github.com/lemonyxk/kitty/json"
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

type Jv struct {
	Event string
	Data  any
}

func (s *Stream[T]) Json(data Jv) error {
	return s.JsonEmit(data.Event, data.Data)
}

type Pv struct {
	Event string
	Data  proto.Message
}

func (s *Stream[T]) ProtoBuf(data Pv) error {
	return s.ProtoBufEmit(data.Event, data.Data)
}

type Rv struct {
	Event string
	Data  []byte
}

func (s *Stream[T]) Send(data Rv) error {
	return s.Emit(data.Event, data.Data)
}
