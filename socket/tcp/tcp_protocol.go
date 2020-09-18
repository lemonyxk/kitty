/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2020-06-06 10:54
**/

package tcp

import (
	"errors"

	"github.com/lemoyxk/kitty/socket"
)

type Protocol interface {
	Decode(message []byte) (version byte, messageType int, protoType int, route []byte, body []byte)
	Encode(route []byte, body []byte, messageType int, protoType int) []byte
	Reader() func(n int, buf []byte, fn func(bytes []byte)) error
}

type DefaultProtocol struct{}

func (p *DefaultProtocol) Decode(message []byte) (version byte, messageType int, protoType int, route []byte, body []byte) {
	if !isHeaderInvalid(message) {
		return 0, 0, 0, nil, nil
	}

	if getLen(message) != len(message) {
		return 0, 0, 0, nil, nil
	}

	return message[0], int(message[1]), int(message[2]), message[8 : 8+message[3]], message[8+message[3]:]
}

func (p *DefaultProtocol) Encode(route []byte, body []byte, messageType int, protoType int) []byte {
	switch messageType {
	case socket.TextData:
		return packText(route, body, protoType)
	case socket.BinData:
		return packBin(route, body, protoType)
	case socket.PingData:
		return []byte{socket.Version, byte(socket.PingData), byte(protoType), 0, 0, 0, 0, 0}
	case socket.PongData:
		return []byte{socket.Version, byte(socket.PongData), byte(protoType), 0, 0, 0, 0, 0}
	}

	return nil
}

func (p *DefaultProtocol) Reader() func(n int, buf []byte, fn func(bytes []byte)) error {

	var singleMessageLen = 0

	var message []byte

	return func(n int, buf []byte, fn func(bytes []byte)) error {

		message = append(message, buf[0:n]...)

		// read continue
		if len(message) < 8 {
			return nil
		}

		for {

			// jump out and read continue
			if len(message) < 8 {
				return nil
			}

			// just begin
			if singleMessageLen == 0 {

				// proto error
				if !isHeaderInvalid(message) {
					// message = message[0:0]
					// singleMessageLen = 0
					return errors.New("invalid header")
				}

				singleMessageLen = getLen(message)
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

func isHeaderInvalid(message []byte) bool {

	if len(message) < 8 {
		return false
	}

	// version
	if message[0] != socket.Version {
		return false
	}

	// message type
	if message[1] != byte(socket.TextData) && message[1] != byte(socket.BinData) && message[1] != byte(socket.PingData) && message[1] != byte(socket.PongData) {
		return false
	}

	// proto type
	if message[2] != byte(socket.Json) && message[2] != byte(socket.ProtoBuf) && message[2] != byte(socket.Text) {
		return false
	}

	return true
}

func convert(message []byte) (a, b, c, d, e int) {
	return int(message[3]), int(message[7]), int(message[6]), int(message[5]), int(message[4])
}

func getLen(message []byte) int {
	var a, b, c, d, e = convert(message[:8])
	if message[1] == byte(socket.TextData) {
		return a + (b | c<<7 | d<<14 | e<<21) + 8
	} else {
		return a + (b | c<<8 | d<<16 | e<<24) + 8
	}
}

func packText(route []byte, body []byte, protoType int) []byte {

	var bl = len(body)

	// data struct
	var data []byte

	// 0 version
	data = append(data, socket.Version)

	// 1 message type
	data = append(data, byte(socket.TextData))

	// 2 proto type
	data = append(data, byte(protoType))

	// 3 route len
	data = append(data, byte(len(route)&0x007f))

	// 4 body len
	data = append(data, byte(bl>>21&0x007f))

	// 5 body len
	data = append(data, byte(bl>>14&0x007f))

	// 6 body len
	data = append(data, byte(bl>>7&0x007f))

	// 7 body len
	data = append(data, byte(bl&0x007f))

	data = append(data, route...)

	data = append(data, body...)

	return data

}
func packBin(route []byte, body []byte, protoType int) []byte {

	var bl = len(body)

	// data struct
	var data []byte

	// 0 version
	data = append(data, socket.Version)

	// 1 message type
	data = append(data, byte(socket.BinData))

	// 2 proto type
	data = append(data, byte(protoType))

	// 3 route len
	data = append(data, byte(len(route)&0x00ff))

	// 4 body len
	data = append(data, byte(bl>>24&0x00ff))

	// 5 body len
	data = append(data, byte(bl>>16&0x00ff))

	// 6 body len
	data = append(data, byte(bl>>8&0x00ff))

	// 7 body len
	data = append(data, byte(bl&0x00ff))

	data = append(data, route...)

	data = append(data, body...)

	return data
}