/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2022-05-23 20:46
**/

package async

import (
	"sync"
	"time"

	"github.com/lemonyxk/kitty/v2/errors"
	"github.com/lemonyxk/kitty/v2/router"
	"github.com/lemonyxk/kitty/v2/socket"
)

type Server[T any] interface {
	Emit(fd int64, pack socket.Pack) error
	JsonEmit(fd int64, pack socket.JsonPack) error
	ProtoBufEmit(fd int64, pack socket.ProtoBufPack) error
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

func (a *asyncServer[T]) Emit(fd int64, pack socket.Pack) (*socket.Stream[T], error) {
	a.mux.Lock()
	defer a.mux.Unlock()

	var ch = make(chan *socket.Stream[T])
	a.server.GetRouter().Route(pack.Event).Handler(func(stream *socket.Stream[T]) error {
		ch <- stream
		return nil
	})

	defer func() { a.server.GetRouter().Remove(pack.Event) }()

	var err = a.server.Emit(fd, pack)
	if err != nil {
		return nil, err
	}

	var timeout = time.After(a.server.GetDailTimeout())

	select {
	case <-timeout:
		return nil, errors.New("timeout")
	case stream := <-ch:
		return stream, nil
	}
}

func (a *asyncServer[T]) JsonEmit(fd int64, pack socket.JsonPack) (*socket.Stream[T], error) {
	a.mux.Lock()
	defer a.mux.Unlock()

	var ch = make(chan *socket.Stream[T])
	a.server.GetRouter().Route(pack.Event).Handler(func(stream *socket.Stream[T]) error {
		ch <- stream
		return nil
	})

	defer func() { a.server.GetRouter().Remove(pack.Event) }()

	var err = a.server.JsonEmit(fd, pack)
	if err != nil {
		return nil, err
	}

	var timeout = time.After(a.server.GetDailTimeout())

	select {
	case <-timeout:
		return nil, errors.New("timeout")
	case stream := <-ch:
		return stream, nil
	}
}

func (a *asyncServer[T]) ProtoBufEmit(fd int64, pack socket.ProtoBufPack) (*socket.Stream[T], error) {
	a.mux.Lock()
	defer a.mux.Unlock()

	var ch = make(chan *socket.Stream[T])
	a.server.GetRouter().Route(pack.Event).Handler(func(stream *socket.Stream[T]) error {
		ch <- stream
		return nil
	})

	defer func() { a.server.GetRouter().Remove(pack.Event) }()

	var err = a.server.ProtoBufEmit(fd, pack)
	if err != nil {
		return nil, err
	}

	var timeout = time.After(a.server.GetDailTimeout())

	select {
	case <-timeout:
		return nil, errors.New("timeout")
	case stream := <-ch:
		return stream, nil
	}
}
