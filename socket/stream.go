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
	"github.com/golang/protobuf/proto"
	jsoniter "github.com/json-iterator/go"
	"github.com/lemonyxk/kitty/v2/kitty"
	"github.com/lemonyxk/kitty/v2/socket/protocol"
)

type Emitter interface {
	JsonEmit(event string, data any) error
	ProtoBufEmit(event string, data proto.Message) error
	Emit(event string, data []byte) error
	Pack(messageType byte, messageID int64, route []byte, body []byte) error
	Push(msg []byte) error
}

func NewStream[T Emitter](conn T, messageID int64, event string, data []byte) *Stream[T] {
	return &Stream[T]{conn: conn, messageID: messageID, Event: event, Data: data}
}

type Stream[T Emitter] struct {
	Data  []byte
	Event string

	Context kitty.Context
	Params  Params
	Logger  kitty.Logger

	conn      T
	messageID int64
}

func (s *Stream[T]) Conn() T {
	return s.conn
}

func (s *Stream[T]) MessageID() int64 {
	return s.messageID
}

func (s *Stream[T]) Emit(event string, data []byte) error {
	return s.conn.Pack(protocol.Bin, s.messageID, []byte(event), data)
}

func (s *Stream[T]) JsonEmit(event string, data any) error {
	msg, err := jsoniter.Marshal(data)
	if err != nil {
		return err
	}
	return s.conn.Pack(protocol.Bin, s.messageID, []byte(event), msg)
}

func (s *Stream[T]) ProtoBufEmit(event string, data proto.Message) error {
	msg, err := proto.Marshal(data)
	if err != nil {
		return err
	}
	return s.conn.Pack(protocol.Bin, s.messageID, []byte(event), msg)
}

func (s *Stream[T]) Push(message []byte) error {
	return s.conn.Push(message)
}
