/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2021-05-13 19:33
**/

package http

import (
	"google.golang.org/protobuf/proto"
	"strconv"
)

type Protobuf struct {
	bts []byte
	t   proto.Message
}

func (p *Protobuf) Reset(v proto.Message) error {
	var bts, err = proto.Marshal(v)
	if err != nil {
		return err
	}
	p.bts = bts
	return err
}

func (p *Protobuf) Bytes() []byte {
	return p.bts
}

func (p *Protobuf) String() string {
	return "<Protobuf: " + strconv.Itoa(len(p.bts)) + " >"
}

func (p *Protobuf) Decode(v proto.Message) error {
	if len(p.bts) == 0 {
		return nil
	}
	p.t = v
	return proto.Unmarshal(p.bts, v)
}

func (p *Protobuf) Encode() ([]byte, error) {
	return proto.Marshal(p.t)
}
