/**
* @program: lemon
*
* @description:
*
* @author: lemon
*
* @create: 2020-06-06 10:54
**/

package protocol

import (
	"encoding/binary"
)

type DefaultWsProtocol struct{}

func (d *DefaultWsProtocol) HeadLen() int {
	return 16
}

func (d *DefaultWsProtocol) Pong() []byte {
	return PongMessage
}

func (d *DefaultWsProtocol) Ping() []byte {
	return PingMessage
}

func (d *DefaultWsProtocol) IsPing(messageType byte) bool {
	return messageType == Ping
}

func (d *DefaultWsProtocol) IsPong(messageType byte) bool {
	return messageType == Pong
}

func (d *DefaultWsProtocol) IsUnknown(messageType byte) bool {
	return messageType == Unknown
}

func (d *DefaultWsProtocol) Decode(message []byte) (messageType byte, id int64, route []byte, body []byte) {
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

func (d *DefaultWsProtocol) Encode(messageType byte, id int64, route []byte, body []byte) []byte {
	switch messageType {
	case Bin:
		return d.packBin(id, route, body)
	case Ping:
		return PingMessage
	case Pong:
		return PongMessage
	}
	return nil
}

func (d *DefaultWsProtocol) Reader() func(n int, buf []byte, fn func(bytes []byte)) error {
	return func(n int, buf []byte, fn func(bytes []byte)) error {
		fn(buf[:n])
		return nil
	}
}

func (d *DefaultWsProtocol) isHeaderInvalid(message []byte) bool {

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
	if message[2] != Bin && message[2] != Ping && message[2] != Pong {
		return false
	}

	return true
}

func (d *DefaultWsProtocol) getLen(message []byte) int {
	headLen := d.HeadLen()

	if len(message) < headLen {
		return 0
	}

	var rl = int(message[3])

	var bl = binary.BigEndian.Uint32(message[4:8])

	return rl + int(bl) + headLen
}

func (d *DefaultWsProtocol) packBin(id int64, route []byte, body []byte) []byte {

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
