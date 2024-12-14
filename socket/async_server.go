/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2022-05-23 20:46
**/

package socket

import (
	json "github.com/lemonyxk/kitty/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lemonyxk/kitty/errors"
	"github.com/lemonyxk/kitty/router"
	"github.com/lemonyxk/kitty/socket/protocol"
	"google.golang.org/protobuf/proto"
)

type server[T Packer, P any] interface {
	Conn(fd int64) (T, error)
	GetDailTimeout() time.Duration
	GetRouter() *router.Router[*Stream[T], P]
}

type Server[T Packer, P any] struct {
	server server[T, P]
	mux    *sync.Mutex
}

func NewAsyncServer[T Packer, P any](server server[T, P]) *Server[T, P] {
	return &Server[T, P]{
		server: server,
		mux:    &sync.Mutex{},
	}
}

func (s *Server[T, P]) Sender(fd int64) (*ServerSender[T, P], error) {
	var conn, err = s.server.Conn(fd)
	if err != nil {
		return nil, err
	}
	return &ServerSender[T, P]{
		sender: &sender[T]{conn: conn, code: 0, messageID: 0},
		server: s.server,
		mux:    s.mux,
	}, nil
}

type ServerSender[T Packer, P any] struct {
	server server[T, P]
	mux    *sync.Mutex
	*sender[T]
}

func (s *ServerSender[T, P]) Emit(event string, data []byte) (*Stream[T], error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	var ch = make(chan *Stream[T])
	s.server.GetRouter().Route(event).Handler(func(stream *Stream[T]) error {
		ch <- stream
		return nil
	})

	defer func() { s.server.GetRouter().Remove(event) }()

	err := s.conn.Pack(s.order, protocol.Bin, s.code, atomic.AddUint64(&s.messageID, 1), []byte(event), data)
	if err != nil {
		return nil, err
	}

	var timeout = time.After(s.server.GetDailTimeout())

	select {
	case <-timeout:
		return nil, errors.Timeout
	case stream := <-ch:
		return stream, nil
	}
}

func (s *ServerSender[T, P]) JsonEmit(event string, data any) (*Stream[T], error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	var ch = make(chan *Stream[T])
	s.server.GetRouter().Route(event).Handler(func(stream *Stream[T]) error {
		ch <- stream
		return nil
	})

	defer func() { s.server.GetRouter().Remove(event) }()

	msg, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	err = s.conn.Pack(s.order, protocol.Json, s.code, atomic.AddUint64(&s.messageID, 1), []byte(event), msg)
	if err != nil {
		return nil, err
	}

	var timeout = time.After(s.server.GetDailTimeout())

	select {
	case <-timeout:
		return nil, errors.Timeout
	case stream := <-ch:
		return stream, nil
	}
}

func (s *ServerSender[T, P]) ProtoBufEmit(event string, data proto.Message) (*Stream[T], error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	var ch = make(chan *Stream[T])
	s.server.GetRouter().Route(event).Handler(func(stream *Stream[T]) error {
		ch <- stream
		return nil
	})

	defer func() { s.server.GetRouter().Remove(event) }()

	msg, err := proto.Marshal(data)
	if err != nil {
		return nil, err
	}

	err = s.conn.Pack(s.order, protocol.ProtoBuf, s.code, atomic.AddUint64(&s.messageID, 1), []byte(event), msg)
	if err != nil {
		return nil, err
	}

	var timeout = time.After(s.server.GetDailTimeout())

	select {
	case <-timeout:
		return nil, errors.Timeout
	case stream := <-ch:
		return stream, nil
	}
}
