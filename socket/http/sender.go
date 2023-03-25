/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2023-02-27 23:41
**/

package http

import (
	"fmt"
	"io"
	"net/http"

	"github.com/json-iterator/go"
	"github.com/lemonyxk/kitty/kitty/header"
)

type Sender struct {
	response http.ResponseWriter
	request  *http.Request
}

func (s *Sender) Respond(code int, msg any) error {
	s.response.WriteHeader(code)
	switch msg.(type) {
	case string:
		return s.String(msg.(string))
	case []byte:
		return s.Bytes(msg.([]byte))
	default:
		return s.Any(msg)
	}
}

func (s *Sender) Any(data any) error {
	var _, err = s.response.Write([]byte(fmt.Sprintf("%+v", data)))
	return err
}

func (s *Sender) Json(data any) error {
	s.response.Header().Set(header.ContentType, header.ApplicationJson)
	bts, err := jsoniter.Marshal(data)
	if err != nil {
		return err
	}
	_, err = s.response.Write(bts)
	return err
}

func (s *Sender) String(data string) error {
	_, err := s.response.Write([]byte(data))
	return err
}

func (s *Sender) Bytes(data []byte) error {
	_, err := s.response.Write(data)
	return err
}

func (s *Sender) Error(err error) error {
	var e = s.String(err.Error())
	if e != nil {
		return e
	}
	return err
}

func (s *Sender) File(fileName string, file io.Reader) error {
	s.response.Header().Set(header.ContentType, header.ApplicationOctetStream)
	s.response.Header().Set(header.ContentDisposition, "attachment;filename="+fileName)
	_, err := io.Copy(s.response, file)
	return err
}
