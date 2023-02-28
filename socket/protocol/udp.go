/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2021-02-12 23:59
**/

package protocol

import (
	"encoding/binary"
)

const (
	Open  byte = 5
	Close byte = 6
)

var OpenMessage = []byte{0x0, 0x0, 0x5, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
var CloseMessage = []byte{0x0, 0x0, 0x6, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}

type UDPProtocol interface {
	Protocol
	GetMessageType([]byte) byte
	PackClose() []byte
	PackOpen() []byte
	IsClose(byte) bool
	IsOpen(byte) bool
}

type DefaultUdpProtocol struct{}

func (d *DefaultUdpProtocol) GetMessageType(message []byte) byte {
	return message[2]
}

func (d *DefaultUdpProtocol) PackPong() []byte {
	return PongMessage
}

func (d *DefaultUdpProtocol) PackPing() []byte {
	return PingMessage
}

func (d *DefaultUdpProtocol) PackClose() []byte {
	return CloseMessage
}

func (d *DefaultUdpProtocol) PackOpen() []byte {
	return OpenMessage
}

func (d *DefaultUdpProtocol) IsPing(messageType byte) bool {
	return messageType == Ping
}

func (d *DefaultUdpProtocol) IsPong(messageType byte) bool {
	return messageType == Pong
}

func (d *DefaultUdpProtocol) IsClose(messageType byte) bool {
	return messageType == Close
}

func (d *DefaultUdpProtocol) IsOpen(messageType byte) bool {
	return messageType == Open
}

func (d *DefaultUdpProtocol) IsUnknown(messageType byte) bool {
	return messageType == Unknown
}

func (d *DefaultUdpProtocol) HeadLen() int {
	return 16
}

func (d *DefaultUdpProtocol) Decode(message []byte) (messageType byte, id int64, route []byte, body []byte) {
	if !d.isHeaderInvalid(message) {
		return 0, 0, nil, nil
	}

	if d.getLen(message) != len(message) {
		return 0, 0, nil, nil
	}

	headLen := d.HeadLen()

	return message[2],
		int64(binary.BigEndian.Uint64(message[8:headLen])),
		message[headLen : headLen+int(message[3])], message[headLen+int(message[3]):]
}

func (d *DefaultUdpProtocol) Encode(messageType byte, id int64, route []byte, body []byte) []byte {
	switch messageType {
	case Bin:
		return d.packBin(id, route, body)
	case Ping:
		return PingMessage
	case Pong:
		return PongMessage
	case Close:
		return CloseMessage
	case Open:
		return OpenMessage
	}
	return nil
}

func (d *DefaultUdpProtocol) Reader() func(n int, buf []byte, fn func(bytes []byte)) error {
	var message []byte
	return func(n int, buf []byte, fn func(bytes []byte)) error {
		message = append(message, buf[0:n]...)
		fn(message)
		message = message[n:]
		return nil
	}
}

func (d *DefaultUdpProtocol) isHeaderInvalid(message []byte) bool {

	if len(message) < d.HeadLen() {
		return false
	}

	// keep
	if message[0] != 0 {
		return false
	}

	// keep
	if message[1] != 0 {
		return false
	}

	// message type
	if message[2] != Bin &&
		message[2] != Ping &&
		message[2] != Pong &&
		message[2] != Open &&
		message[2] != Close {
		return false
	}

	return true
}

func (d *DefaultUdpProtocol) getLen(message []byte) int {
	headLen := d.HeadLen()

	if len(message) < headLen {
		return 0
	}

	var rl = int(message[3])

	var bl = binary.BigEndian.Uint32(message[4:8])

	return rl + int(bl) + headLen
}

func (d *DefaultUdpProtocol) packBin(id int64, route []byte, body []byte) []byte {

	var rl = len(route)

	var bl = len(body)

	// data struct
	headLen := d.HeadLen()

	var data = make([]byte, headLen+rl+bl)

	// 0 keep
	data[0] = 0

	// 1 keep
	data[1] = 0

	// 2 message type
	data[2] = Bin

	// 3 route len
	data[3] = byte(rl)

	// 4 - 7 body len
	binary.BigEndian.PutUint32(data[4:8], uint32(bl))

	// 8 - 15 id
	binary.BigEndian.PutUint64(data[8:headLen], uint64(id))

	copy(data[headLen:headLen+rl], route)

	copy(data[headLen+rl:headLen+rl+bl], body)

	return data
}
