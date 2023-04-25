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
	"sync"
	"time"

	"github.com/lemonyxk/kitty/errors"
	"github.com/lemonyxk/kitty/router"
	"google.golang.org/protobuf/proto"
)

type server[T Packer] interface {
	Conn(fd int64) (T, error)
	GetDailTimeout() time.Duration
	GetRouter() *router.Router[*Stream[T]]
}

type Server[T Packer] struct {
	server server[T]
	mux    *sync.Mutex
}

func NewAsyncServer[T Packer](server server[T]) *Server[T] {
	return &Server[T]{
		server: server,
		mux:    &sync.Mutex{},
	}
}

func (s *Server[T]) Sender(fd int64) (*ServerSender[T], error) {
	var conn, err = s.server.Conn(fd)
	if err != nil {
		return nil, err
	}
	return &ServerSender[T]{
		sender: &sender[T]{conn: conn, code: 0, messageID: 0},
		server: s.server,
		mux:    s.mux,
	}, nil
}

type ServerSender[T Packer] struct {
	server server[T]
	mux    *sync.Mutex
	*sender[T]
}

func (s *ServerSender[T]) Emit(event string, data []byte) (*Stream[T], error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	var ch = make(chan *Stream[T])
	s.server.GetRouter().Route(event).Handler(func(stream *Stream[T]) error {
		ch <- stream
		return nil
	})

	defer func() { s.server.GetRouter().Remove(event) }()

	err := s.sender.Emit(event, data)
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

func (s *ServerSender[T]) JsonEmit(event string, data any) (*Stream[T], error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	var ch = make(chan *Stream[T])
	s.server.GetRouter().Route(event).Handler(func(stream *Stream[T]) error {
		ch <- stream
		return nil
	})

	defer func() { s.server.GetRouter().Remove(event) }()

	err := s.sender.JsonEmit(event, data)
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

func (s *ServerSender[T]) ProtoBufEmit(event string, data proto.Message) (*Stream[T], error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	var ch = make(chan *Stream[T])
	s.server.GetRouter().Route(event).Handler(func(stream *Stream[T]) error {
		ch <- stream
		return nil
	})

	defer func() { s.server.GetRouter().Remove(event) }()

	err := s.sender.ProtoBufEmit(event, data)
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
