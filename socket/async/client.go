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

	"github.com/lemonyxk/kitty/v2/errors"
	"github.com/lemonyxk/kitty/v2/router"
	"github.com/lemonyxk/kitty/v2/socket"
)

type Client[T any] interface {
	JsonEmit(pack socket.JsonPack) error
	ProtoBufEmit(pack socket.ProtoBufPack) error
	Emit(pack socket.Pack) error
	GetRouter() *router.Router[*socket.Stream[T]]
	GetDailTimeout() time.Duration
}

type asyncClient[T any] struct {
	client Client[T]
	mux    sync.Mutex
}

func NewClient[T any](client Client[T]) *asyncClient[T] {
	return &asyncClient[T]{client: client}
}

func (a *asyncClient[T]) Emit(pack socket.Pack) (*socket.Stream[T], error) {
	a.mux.Lock()
	defer a.mux.Unlock()

	var ch = make(chan *socket.Stream[T])
	a.client.GetRouter().Route(pack.Event).Handler(func(stream *socket.Stream[T]) error {
		ch <- stream
		return nil
	})

	defer func() { a.client.GetRouter().Remove(pack.Event) }()

	var err = a.client.Emit(pack)
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

func (a *asyncClient[T]) JsonEmit(pack socket.JsonPack) (*socket.Stream[T], error) {
	a.mux.Lock()
	defer a.mux.Unlock()

	var ch = make(chan *socket.Stream[T])
	a.client.GetRouter().Route(pack.Event).Handler(func(stream *socket.Stream[T]) error {
		ch <- stream
		return nil
	})

	defer func() { a.client.GetRouter().Remove(pack.Event) }()

	var err = a.client.JsonEmit(pack)
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

func (a *asyncClient[T]) ProtoBufEmit(pack socket.ProtoBufPack) (*socket.Stream[T], error) {
	a.mux.Lock()
	defer a.mux.Unlock()

	var ch = make(chan *socket.Stream[T])
	a.client.GetRouter().Route(pack.Event).Handler(func(stream *socket.Stream[T]) error {
		ch <- stream
		return nil
	})

	defer func() { a.client.GetRouter().Remove(pack.Event) }()

	var err = a.client.ProtoBufEmit(pack)
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
