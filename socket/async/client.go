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

	jsoniter "github.com/json-iterator/go"
	"github.com/lemonyxk/kitty/errors"
	"github.com/lemonyxk/kitty/router"
	"github.com/lemonyxk/kitty/socket"
	"github.com/lemonyxk/kitty/socket/protocol"
	"google.golang.org/protobuf/proto"
)

type Client[T socket.Packer] interface {
	Conn() T
	GetRouter() *router.Router[*socket.Stream[T]]
	GetDailTimeout() time.Duration
}

type asyncClient[T socket.Packer] struct {
	client Client[T]
	mux    sync.Mutex
}

func NewClient[T socket.Packer](client Client[T]) *asyncClient[T] {
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

	var err = a.client.Conn().Pack(protocol.Bin, 0, 0, []byte(event), data)
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

	msg, err := jsoniter.Marshal(data)
	if err != nil {
		return nil, err
	}
	err = a.client.Conn().Pack(protocol.Bin, 0, 0, []byte(event), msg)
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

	msg, err := proto.Marshal(data)
	if err != nil {
		return nil, err
	}

	err = a.client.Conn().Pack(protocol.Bin, 0, 0, []byte(event), msg)
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
