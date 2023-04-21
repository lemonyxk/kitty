/**
* @program: lemon
*
* @description:
*
* @author: lemon
*
* @create: 2020-06-06 10:54
**/

package main

import (
	"bytes"

	"github.com/lemonyxk/kitty/errors"
	"github.com/lemonyxk/kitty/socket/protocol"
)

type CustomTcp struct{}

// 0 messageType
// : split route and body
// \r\n end
// min length: 4
// example:
// 2hello:world\r\n

var CustomPingMessage = []byte{9, ':', '\r', '\n'}
var CustomPongMessage = []byte{10, ':', '\r', '\n'}

func (d *CustomTcp) HeadLen() int {
	return 4
}

func (d *CustomTcp) PackPong() []byte {
	return CustomPongMessage
}

func (d *CustomTcp) PackPing() []byte {
	return CustomPingMessage
}

func (d *CustomTcp) IsPing(messageType byte) bool {
	return messageType == protocol.Ping
}

func (d *CustomTcp) IsPong(messageType byte) bool {
	return messageType == protocol.Pong
}

func (d *CustomTcp) IsUnknown(messageType byte) bool {
	return messageType == protocol.Unknown
}

func (d *CustomTcp) Decode(message []byte) (messageType byte, code int, id int64, route []byte, body []byte) {
	if !d.isHeaderInvalid(message) {
		return 0, 0, 0, nil, nil
	}

	var index = bytes.IndexByte(message, ':')
	if index == -1 {
		return 0, 0, 0, nil, nil
	}

	messageType = message[0]
	id = 0
	code = 0
	route = message[1:index]
	body = message[index+1:]

	return messageType, code, id, route, body
}

func (d *CustomTcp) Encode(messageType byte, code int, id int64, route []byte, body []byte) []byte {
	switch messageType {
	case protocol.Bin:
		return d.packBin(id, route, body)
	case protocol.Ping:
		return CustomPingMessage
	case protocol.Pong:
		return CustomPongMessage
	}
	return nil
}

func (d *CustomTcp) Reader() func(n int, buf []byte, fn func(bytes []byte)) error {

	var message []byte

	return func(n int, buf []byte, fn func(bytes []byte)) error {

		message = append(message, buf[0:n]...)

		if len(message) < d.HeadLen() {
			return nil
		}

		for {

			if len(message) < d.HeadLen() {
				return nil
			}

			if !d.isHeaderInvalid(message) {
				message = message[0:0]
				return errors.Invalid
			}

			var index = bytes.Index(message, []byte("\r\n"))

			if index == -1 {
				return nil
			}

			fn(message[:index])

			message = message[index+2:]

		}

	}
}

func (d *CustomTcp) isHeaderInvalid(message []byte) bool {
	if len(message) < d.HeadLen() {
		return false
	}

	// message type
	if message[0] != protocol.Bin && message[0] != protocol.Ping && message[0] != protocol.Pong {
		return false
	}

	return true
}

func (d *CustomTcp) packBin(id int64, route []byte, body []byte) []byte {

	// messageType route:body

	var rl = len(route)

	var bl = len(body)

	var data = make([]byte, d.HeadLen()+rl+bl)

	data[0] = protocol.Bin

	copy(data[1:1+rl], route)

	copy(data[1+rl:], ":")

	copy(data[2+rl:2+rl+bl], body)

	copy(data[2+rl+bl:], "\r\n")

	return data
}
