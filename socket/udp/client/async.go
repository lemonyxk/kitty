/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2021-03-03 15:26
**/

package client

import (
	"errors"
	"sync"
	"time"

	"github.com/lemoyxk/kitty/v2/socket"
)

var async = new(Async)

type Async struct {
	client *Client
	mux    sync.Mutex
}

func (a *Async) Emit(pack socket.Pack) (*socket.Stream, error) {
	a.mux.Lock()
	defer a.mux.Unlock()

	var ch = make(chan *socket.Stream)
	a.client.GetRouter().Route(pack.Event).Handler(func(client *Client, stream *socket.Stream) error {
		ch <- stream
		return nil
	})

	defer func() { a.client.GetRouter().Remove(pack.Event) }()

	var err = a.client.Emit(pack)
	if err != nil {
		return nil, err
	}

	var timeout = time.After(a.client.DailTimeout)

	select {
	case <-timeout:
		return nil, errors.New("timeout")
	case stream := <-ch:
		return stream, nil
	}
}

func (a *Async) JsonEmit(pack socket.JsonPack) (*socket.Stream, error) {
	a.mux.Lock()
	defer a.mux.Unlock()

	var ch = make(chan *socket.Stream)
	a.client.GetRouter().Route(pack.Event).Handler(func(client *Client, stream *socket.Stream) error {
		ch <- stream
		return nil
	})

	defer func() { a.client.GetRouter().Remove(pack.Event) }()

	var err = a.client.JsonEmit(pack)
	if err != nil {
		return nil, err
	}

	var timeout = time.After(a.client.DailTimeout)

	select {
	case <-timeout:
		return nil, errors.New("timeout")
	case stream := <-ch:
		return stream, nil
	}
}

func (a *Async) ProtoBufEmit(pack socket.ProtoBufPack) (*socket.Stream, error) {
	a.mux.Lock()
	defer a.mux.Unlock()

	var ch = make(chan *socket.Stream)
	a.client.GetRouter().Route(pack.Event).Handler(func(client *Client, stream *socket.Stream) error {
		ch <- stream
		return nil
	})

	defer func() { a.client.GetRouter().Remove(pack.Event) }()

	var err = a.client.ProtoBufEmit(pack)
	if err != nil {
		return nil, err
	}

	var timeout = time.After(a.client.DailTimeout)

	select {
	case <-timeout:
		return nil, errors.New("timeout")
	case stream := <-ch:
		return stream, nil
	}
}

func (c *Client) Async() *Async {
	async.client = c
	return async
}
