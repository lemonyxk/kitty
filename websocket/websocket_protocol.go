/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2020-06-06 10:54
**/

package websocket

import (
	"github.com/Lemo-yxk/lemo"
)

type Protocol interface {
	Decode(message []byte) (version byte, messageType int, protoType int, route []byte, body []byte)
	Encode(route []byte, body []byte, messageType int, protoType int) []byte
	Read()
}

type DefaultProtocol struct{}

func (Protocol *DefaultProtocol) Decode(message []byte) (version byte, messageType int, protoType int, route []byte, body []byte) {
	if !isHeaderInvalid(message) {
		route, body := parseMessage(message)

		if len(route) != 0 {
			return lemo.Version, lemo.Text, lemo.Json, route, body
		}
		return 0, 0, 0, nil, nil
	}

	if getLen(message) != len(message) {
		return 0, 0, 0, nil, nil
	}

	return message[0], int(message[1]), int(message[2]), message[8 : 8+message[3]], message[8+message[3]:]
}

func (Protocol *DefaultProtocol) Encode(route []byte, body []byte, messageType int, protoType int) []byte {
	switch messageType {
	case lemo.TextData:
		return packText(route, body, protoType)
	case lemo.BinData:
		return packBin(route, body, protoType)
	case lemo.PingData:
		return []byte{lemo.Version, byte(lemo.PingData), byte(protoType), 0, 0, 0, 0, 0}
	case lemo.PongData:
		return []byte{lemo.Version, byte(lemo.PongData), byte(protoType), 0, 0, 0, 0, 0}
	}

	return nil
}

func (Protocol *DefaultProtocol) Read() {

}

func isHeaderInvalid(message []byte) bool {

	if len(message) < 8 {
		return false
	}

	// version
	if message[0] != lemo.Version {
		return false
	}

	// message type
	if message[1] != byte(lemo.TextData) && message[1] != byte(lemo.BinData) && message[1] != byte(lemo.PingData) && message[1] != byte(lemo.PongData) {
		return false
	}

	// proto type
	if message[2] != byte(lemo.Json) && message[2] != byte(lemo.ProtoBuf) && message[2] != byte(lemo.Text) {
		return false
	}

	return true
}

func convert(message []byte) (a, b, c, d, e int) {
	return int(message[3]), int(message[7]), int(message[6]), int(message[5]), int(message[4])
}

func getLen(message []byte) int {
	var a, b, c, d, e = convert(message[:8])
	if message[1] == byte(lemo.TextData) {
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
	data = append(data, lemo.Version)

	// 1 message type
	data = append(data, byte(lemo.TextData))

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
	data = append(data, lemo.Version)

	// 1 message type
	data = append(data, byte(lemo.BinData))

	// 2 proto type
	data = append(data, byte(protoType))

	// 3 route len
	data = append(data, byte(len(route)&0x00ff))

	// 4 body len
	data = append(data, byte(bl>>28&0x00ff))

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

func parseMessage(bts []byte) ([]byte, []byte) {

	var s, e int

	var l = len(bts)

	if l < 9 {
		return nil, nil
	}

	// 正序
	if bts[8] == 58 {

		s = 8

		for i := 0; i < len(bts); i++ {
			if bts[i] == 44 {
				e = i
				break
			}
		}

		if e == 0 {
			return bts[s+2 : l-2], nil
		}

		return bts[s+2 : e-1], bts[e+8 : l-1]

	} else {

		for i := l - 1; i >= 0; i-- {

			if bts[i] == 58 {
				s = i
			}

			if bts[i] == 44 {
				e = i
				break
			}
		}

		if s == 0 {
			return nil, nil
		}

		if e == 0 {
			return bts[s+2 : l-2], nil
		}

		return bts[s+2 : l-2], bts[8:e]
	}
}
