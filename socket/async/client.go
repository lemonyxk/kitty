/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2022-05-23 19:29
**/

package async

import (
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/lemonyxk/kitty/errors"
	"github.com/lemonyxk/kitty/router"
	"github.com/lemonyxk/kitty/socket"
)

type Client[T socket.Emitter] interface {
	Conn() T
	GetRouter() *router.Router[*socket.Stream[T]]
	GetDailTimeout() time.Duration
}

type asyncClient[T socket.Emitter] struct {
	client Client[T]
	mux    sync.Mutex
}

func NewClient[T socket.Emitter](client Client[T]) *asyncClient[T] {
	return &asyncClient[T]{client: client}
}

func (a *asyncClient[T]) Emit(event string, data []byte) (*socket.Stream[T], error) {
	a.mux.Lock()
	defer a.mux.Unlock()

	var ch = make(chan *socket.Stream[T])
	a.client.GetRouter().Route(event).Handler(func(stream *socket.Stream[T]) error {
		ch <- stream
		return nil
	})

	defer func() { a.client.GetRouter().Remove(event) }()

	var err = a.client.Conn().Emit(event, data)
	if err != nil {
		return nil, err
	}

	var timeout = time.After(a.client.GetDailTimeout())

	select {
	case <-timeout:
		return nil, errors.Timeout
	case stream := <-ch:
		return stream, nil
	}
}

func (a *asyncClient[T]) JsonEmit(event string, data any) (*socket.Stream[T], error) {
	a.mux.Lock()
	defer a.mux.Unlock()

	var ch = make(chan *socket.Stream[T])
	a.client.GetRouter().Route(event).Handler(func(stream *socket.Stream[T]) error {
		ch <- stream
		return nil
	})

	defer func() { a.client.GetRouter().Remove(event) }()

	var err = a.client.Conn().JsonEmit(event, data)
	if err != nil {
		return nil, err
	}

	var timeout = time.After(a.client.GetDailTimeout())

	select {
	case <-timeout:
		return nil, errors.Timeout
	case stream := <-ch:
		return stream, nil
	}
}

func (a *asyncClient[T]) ProtoBufEmit(event string, data proto.Message) (*socket.Stream[T], error) {
	a.mux.Lock()
	defer a.mux.Unlock()

	var ch = make(chan *socket.Stream[T])
	a.client.GetRouter().Route(event).Handler(func(stream *socket.Stream[T]) error {
		ch <- stream
		return nil
	})

	defer func() { a.client.GetRouter().Remove(event) }()

	var err = a.client.Conn().ProtoBufEmit(event, data)
	if err != nil {
		return nil, err
	}

	var timeout = time.After(a.client.GetDailTimeout())

	select {
	case <-timeout:
		return nil, errors.Timeout
	case stream := <-ch:
		return stream, nil
	}
}
