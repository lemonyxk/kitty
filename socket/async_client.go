/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2022-05-23 19:29
**/

package socket

import (
	"sync"
	"time"

	"github.com/lemonyxk/kitty/errors"
	"github.com/lemonyxk/kitty/router"
	"google.golang.org/protobuf/proto"
)

type asyncClient[T Packer] interface {
	Conn() T
	GetRouter() *router.Router[*Stream[T]]
	GetDailTimeout() time.Duration
}

type AsyncClient[T Packer] struct {
	client asyncClient[T]
	mux    sync.Mutex
	*sender[T]
}

func NewAsyncClient[T Packer](client asyncClient[T]) *AsyncClient[T] {
	return &AsyncClient[T]{
		sender: &sender[T]{conn: client.Conn(), code: 0, messageID: 0},
		client: client,
	}
}

func (c *AsyncClient[T]) Emit(event string, data []byte) (*Stream[T], error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	var ch = make(chan *Stream[T])
	c.client.GetRouter().Route(event).Handler(func(stream *Stream[T]) error {
		ch <- stream
		return nil
	})

	defer func() { c.client.GetRouter().Remove(event) }()

	var err = c.sender.Emit(event, data)
	if err != nil {
		return nil, err
	}

	var timeout = time.After(c.client.GetDailTimeout())

	select {
	case <-timeout:
		return nil, errors.Timeout
	case stream := <-ch:
		return stream, nil
	}
}

func (c *AsyncClient[T]) JsonEmit(event string, data any) (*Stream[T], error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	var ch = make(chan *Stream[T])
	c.client.GetRouter().Route(event).Handler(func(stream *Stream[T]) error {
		ch <- stream
		return nil
	})

	defer func() { c.client.GetRouter().Remove(event) }()

	err := c.sender.JsonEmit(event, data)
	if err != nil {
		return nil, err
	}

	var timeout = time.After(c.client.GetDailTimeout())

	select {
	case <-timeout:
		return nil, errors.Timeout
	case stream := <-ch:
		return stream, nil
	}
}

func (c *AsyncClient[T]) ProtoBufEmit(event string, data proto.Message) (*Stream[T], error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	var ch = make(chan *Stream[T])
	c.client.GetRouter().Route(event).Handler(func(stream *Stream[T]) error {
		ch <- stream
		return nil
	})

	defer func() { c.client.GetRouter().Remove(event) }()

	err := c.sender.ProtoBufEmit(event, data)
	if err != nil {
		return nil, err
	}

	var timeout = time.After(c.client.GetDailTimeout())

	select {
	case <-timeout:
		return nil, errors.Timeout
	case stream := <-ch:
		return stream, nil
	}
}
