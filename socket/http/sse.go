/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2023-09-04 12:20
**/

package http

import (
	"bytes"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/lemonyxk/kitty/errors"
	http2 "net/http"
	"sync"
	"sync/atomic"
	"time"
)

type SseConfig struct {
	Retry time.Duration
}

type Sse[T any] struct {
	Stream      *Stream[T]
	LasTEventID int64

	mux     sync.RWMutex
	close   chan struct{}
	isClose bool
}

func (s *Sse[T]) Flush() {
	if f, ok := s.Stream.Response.(http2.Flusher); ok {
		f.Flush()
	}
}

func (s *Sse[T]) String(data string) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	var buf bytes.Buffer
	if s.isClose {
		return errors.New("Write called after Handler finished")
	}
	atomic.AddInt64(&s.LasTEventID, 1)
	buf.WriteString("id: ")
	buf.WriteString(fmt.Sprintf("%d\n", s.LasTEventID))
	buf.WriteString("data: ")
	buf.WriteString(data)
	buf.WriteString("\n\n")
	defer s.Flush()
	return s.Stream.Sender.String(buf.String())
}

func (s *Sse[T]) Bytes(data any) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	var buf bytes.Buffer
	if s.isClose {
		return errors.New("Write called after Handler finished")
	}
	atomic.AddInt64(&s.LasTEventID, 1)
	buf.WriteString("id: ")
	buf.WriteString(fmt.Sprintf("%d\n", s.LasTEventID))
	buf.WriteString("data: ")
	buf.Write(data.([]byte))
	buf.WriteString("\n\n")
	defer s.Flush()
	return s.Stream.Sender.Bytes(buf.Bytes())
}

func (s *Sse[T]) Json(data any) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.isClose {
		return errors.New("Write called after Handler finished")
	}
	var bts, err = jsoniter.Marshal(data)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	atomic.AddInt64(&s.LasTEventID, 1)
	buf.WriteString("id: ")
	buf.WriteString(fmt.Sprintf("%d\n", s.LasTEventID))
	buf.WriteString("data: ")
	buf.Write(bts)
	buf.WriteString("\n\n")
	defer s.Flush()
	return s.Stream.Sender.Bytes(buf.Bytes())
}

func (s *Sse[T]) Any(data any) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	var buf bytes.Buffer
	if s.isClose {
		return errors.New("Write called after Handler finished")
	}
	atomic.AddInt64(&s.LasTEventID, 1)
	buf.WriteString("id: ")
	buf.WriteString(fmt.Sprintf("%d\n", s.LasTEventID))
	buf.WriteString("data: ")
	buf.WriteString(fmt.Sprintf("%+v", data))
	buf.WriteString("\n\n")
	defer s.Flush()
	return s.Stream.Sender.Bytes(buf.Bytes())
}

func (s *Sse[T]) Done() <-chan struct{} {
	return s.close
}

func (s *Sse[T]) Wait() error {
	<-s.Stream.Request.Context().Done()
	s.mux.Lock()
	defer s.mux.Unlock()
	s.close <- struct{}{}
	s.isClose = true
	return nil
}

func (s *Sse[T]) IsClose() bool {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.isClose
}
