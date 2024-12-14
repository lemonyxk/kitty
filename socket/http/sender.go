/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2023-02-27 23:41
**/

package http

import (
	"fmt"
	json "github.com/lemonyxk/kitty/json"
	"io"
	"net/http"

	"github.com/lemonyxk/kitty/kitty/header"
	"google.golang.org/protobuf/proto"
)

func NewSender(w http.ResponseWriter, r *http.Request) *Sender {
	return &Sender{response: w, request: r}
}

type Sender struct {
	response http.ResponseWriter
	request  *http.Request
}

func (s *Sender) Respond(code int, msg any) error {
	s.response.WriteHeader(code)
	if msg == nil {
		return nil
	}
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
	var res []byte
	if data != nil {
		res = []byte(fmt.Sprintf("%+v", data))
	}
	var _, err = s.response.Write(res)
	return err
}

func (s *Sender) Json(data any) error {
	s.response.Header().Set(header.ContentType, header.ApplicationJson)
	var bts, err = json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = s.response.Write(bts)
	return err
	// will have a \n at the end
	//return json.NewEncoder(s.response).Encode(data)
}

func (s *Sender) Protobuf(data proto.Message) error {
	s.response.Header().Set(header.ContentType, header.ApplicationProtobuf)
	bts, err := proto.Marshal(data)
	if err != nil {
		return err
	}
	_, err = s.response.Write(bts)
	return err
}

func (s *Sender) String(data string) error {
	return s.Bytes([]byte(data))
}

func (s *Sender) Bytes(data []byte) error {
	_, err := s.response.Write(data)
	return err
}

func (s *Sender) Redirect(url string, code int) {
	http.Redirect(s.response, s.request, url, code)
}

func (s *Sender) Write(r io.Reader) error {
	_, err := io.Copy(s.response, r)
	return err
}

func (s *Sender) RespondWithError(code int, err error) error {
	var e = s.Respond(code, err.Error())
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
