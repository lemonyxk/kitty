/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2022-05-23 20:46
**/

package async

import (
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/lemonyxk/kitty/v2/errors"
	"github.com/lemonyxk/kitty/v2/router"
	"github.com/lemonyxk/kitty/v2/socket"
)

type Server[T any] interface {
	JsonEmit(fd int64, event string, data any) error
	ProtoBufEmit(fd int64, event string, data proto.Message) error
	Emit(fd int64, event string, data []byte) error
	GetDailTimeout() time.Duration
	GetRouter() *router.Router[*socket.Stream[T]]
}

type asyncServer[T any] struct {
	server Server[T]
	mux    sync.Mutex
}

func NewServer[T any](server Server[T]) *asyncServer[T] {
	return &asyncServer[T]{server: server}
}

func (a *asyncServer[T]) Emit(fd int64, event string, data []byte) (*socket.Stream[T], error) {
	a.mux.Lock()
	defer a.mux.Unlock()

	var ch = make(chan *socket.Stream[T])
	a.server.GetRouter().Route(event).Handler(func(stream *socket.Stream[T]) error {
		ch <- stream
		return nil
	})

	defer func() { a.server.GetRouter().Remove(event) }()

	var err = a.server.Emit(fd, event, data)
	if err != nil {
		return nil, err
	}

	var timeout = time.After(a.server.GetDailTimeout())

	select {
	case <-timeout:
		return nil, errors.Timeout
	case stream := <-ch:
		return stream, nil
	}
}

func (a *asyncServer[T]) JsonEmit(fd int64, event string, data any) (*socket.Stream[T], error) {
	a.mux.Lock()
	defer a.mux.Unlock()

	var ch = make(chan *socket.Stream[T])
	a.server.GetRouter().Route(event).Handler(func(stream *socket.Stream[T]) error {
		ch <- stream
		return nil
	})

	defer func() { a.server.GetRouter().Remove(event) }()

	var err = a.server.JsonEmit(fd, event, data)
	if err != nil {
		return nil, err
	}

	var timeout = time.After(a.server.GetDailTimeout())

	select {
	case <-timeout:
		return nil, errors.Timeout
	case stream := <-ch:
		return stream, nil
	}
}

func (a *asyncServer[T]) ProtoBufEmit(fd int64, event string, data proto.Message) (*socket.Stream[T], error) {
	a.mux.Lock()
	defer a.mux.Unlock()

	var ch = make(chan *socket.Stream[T])
	a.server.GetRouter().Route(event).Handler(func(stream *socket.Stream[T]) error {
		ch <- stream
		return nil
	})

	defer func() { a.server.GetRouter().Remove(event) }()

	var err = a.server.ProtoBufEmit(fd, event, data)
	if err != nil {
		return nil, err
	}

	var timeout = time.After(a.server.GetDailTimeout())

	select {
	case <-timeout:
		return nil, errors.Timeout
	case stream := <-ch:
		return stream, nil
	}
}
