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

	"github.com/lemonyxk/kitty/errors"
)

type DefaultTcpProtocol struct{}

func (d *DefaultTcpProtocol) HeadLen() int {
	return 20
}

func (d *DefaultTcpProtocol) PackPong() []byte {
	return PongMessage
}

func (d *DefaultTcpProtocol) PackPing() []byte {
	return PingMessage
}

func (d *DefaultTcpProtocol) IsPing(messageType byte) bool {
	return messageType == Ping
}

func (d *DefaultTcpProtocol) IsPong(messageType byte) bool {
	return messageType == Pong
}

func (d *DefaultTcpProtocol) IsUnknown(messageType byte) bool {
	return messageType == Unknown
}

func (d *DefaultTcpProtocol) Decode(message []byte) (messageType byte, code uint32, id uint64, route []byte, body []byte) {
	if !d.isHeaderInvalid(message) {
		return 0, 0, 0, nil, nil
	}

	if d.getLen(message) != len(message) {
		return 0, 0, 0, nil, nil
	}

	headLen := d.HeadLen()

	return message[2],
		binary.BigEndian.Uint32(message[8:12]),
		binary.BigEndian.Uint64(message[12:headLen]),
		message[headLen : headLen+int(message[3])], message[headLen+int(message[3]):]
}

func (d *DefaultTcpProtocol) Encode(messageType byte, code uint32, id uint64, route []byte, body []byte) []byte {
	switch messageType {
	case Bin:
		return d.packBin(code, id, route, body)
	case Ping:
		return PingMessage
	case Pong:
		return PongMessage
	}
	return nil
}

func (d *DefaultTcpProtocol) Reader() func(n int, buf []byte, fn func(bytes []byte)) error {

	var singleMessageLen = 0

	var message []byte

	headLen := d.HeadLen()

	return func(n int, buf []byte, fn func(bytes []byte)) error {

		message = append(message, buf[0:n]...)

		// read continue
		if len(message) < headLen {
			return nil
		}

		for {

			// jump out and read continue
			if len(message) < headLen {
				return nil
			}

			// just begin
			if singleMessageLen == 0 {

				// proto error
				if !d.isHeaderInvalid(message) {
					message = message[0:0]
					singleMessageLen = 0
					return errors.Invalid
				}

				singleMessageLen = d.getLen(message)
			}

			// jump out and read continue
			if len(message) < singleMessageLen {
				return nil
			}

			// a complete message
			fn(message[0:singleMessageLen])

			// delete this message
			message = message[singleMessageLen:]

			// reset len
			singleMessageLen = 0
		}

	}
}

func (d *DefaultTcpProtocol) isHeaderInvalid(message []byte) bool {

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

func (d *DefaultTcpProtocol) getLen(message []byte) int {
	headLen := d.HeadLen()

	if len(message) < headLen {
		return 0
	}

	var rl = int(message[3])

	var bl = binary.BigEndian.Uint32(message[4:8])

	return rl + int(bl) + headLen
}

func (d *DefaultTcpProtocol) packBin(code uint32, id uint64, route []byte, body []byte) []byte {

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

	// 8 - 11 code
	binary.BigEndian.PutUint32(data[8:12], code)

	// 12 - 19 id
	binary.BigEndian.PutUint64(data[12:headLen], id)

	copy(data[headLen:headLen+rl], route)

	copy(data[headLen+rl:headLen+rl+bl], body)

	return data
}
