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

	"github.com/lemonyxk/kitty/errors"
	"github.com/lemonyxk/kitty/router"
	"github.com/lemonyxk/kitty/socket"
	"google.golang.org/protobuf/proto"
)

type Server[T socket.Emitter] interface {
	Conn(fd int64) (T, error)
	GetDailTimeout() time.Duration
	GetRouter() *router.Router[*socket.Stream[T]]
}

type asyncServer[T socket.Emitter] struct {
	server Server[T]
	mux    sync.Mutex
}

func NewServer[T socket.Emitter](server Server[T]) *asyncServer[T] {
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

	var conn, err = a.server.Conn(fd)
	if err != nil {
		return nil, err
	}

	err = conn.Emit(event, data)
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

	var conn, err = a.server.Conn(fd)
	if err != nil {
		return nil, err
	}

	err = conn.JsonEmit(event, data)
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

	var conn, err = a.server.Conn(fd)
	if err != nil {
		return nil, err
	}

	err = conn.ProtoBufEmit(event, data)
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
