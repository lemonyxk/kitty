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
	"sync/atomic"
	"time"

	json "github.com/bytedance/sonic"
	"github.com/lemonyxk/kitty/errors"
	"github.com/lemonyxk/kitty/router"
	"github.com/lemonyxk/kitty/socket/protocol"
	"google.golang.org/protobuf/proto"
)

type asyncClient[T Packer, P any] interface {
	Conn() T
	GetRouter() *router.Router[*Stream[T], P]
	GetDailTimeout() time.Duration
}

type AsyncClient[T Packer, P any] struct {
	client asyncClient[T, P]
	mux    sync.Mutex
	*sender[T]
}

func NewAsyncClient[T Packer, P any](client asyncClient[T, P]) *AsyncClient[T, P] {
	return &AsyncClient[T, P]{
		sender: &sender[T]{conn: client.Conn(), code: 0, messageID: 0},
		client: client,
	}
}

func (c *AsyncClient[T, P]) Emit(event string, data []byte) (*Stream[T], error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	var ch = make(chan *Stream[T])
	c.client.GetRouter().Route(event).Handler(func(stream *Stream[T]) error {
		ch <- stream
		return nil
	})

	defer func() { c.client.GetRouter().Remove(event) }()

	var err = c.conn.Pack(c.order, protocol.Bin, c.code, atomic.AddUint64(&c.messageID, 1), []byte(event), data)
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

func (c *AsyncClient[T, P]) JsonEmit(event string, data any) (*Stream[T], error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	var ch = make(chan *Stream[T])
	c.client.GetRouter().Route(event).Handler(func(stream *Stream[T]) error {
		ch <- stream
		return nil
	})

	defer func() { c.client.GetRouter().Remove(event) }()

	msg, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	err = c.conn.Pack(c.order, protocol.Json, c.code, atomic.AddUint64(&c.messageID, 1), []byte(event), msg)
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

func (c *AsyncClient[T, P]) ProtoBufEmit(event string, data proto.Message) (*Stream[T], error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	var ch = make(chan *Stream[T])
	c.client.GetRouter().Route(event).Handler(func(stream *Stream[T]) error {
		ch <- stream
		return nil
	})

	defer func() { c.client.GetRouter().Remove(event) }()

	msg, err := proto.Marshal(data)
	if err != nil {
		return nil, err
	}

	err = c.conn.Pack(c.order, protocol.ProtoBuf, c.code, atomic.AddUint64(&c.messageID, 1), []byte(event), msg)
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
