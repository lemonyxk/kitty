/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-10-09 16:02
**/

package protocol

// 0 version
// 1 message type
// 2 proto type
// 3 route len
// 4 body len
// 5 body len
// 6 body len
// 7 body len

const (
	// Version
	Version byte = 'V'

	// message type
	TextData int = 1
	BinData  int = 2
	PingData int = 9
	PongData int = 10

	// proto type
	Text     int = 1
	Json     int = 2
	ProtoBuf int = 3
)

func Pack(route []byte, body []byte, messageType int, protoType int) []byte {

	switch messageType {
	case TextData:
		return packText(route, body, protoType)
	case BinData:
		return packBin(route, body, protoType)
	case PingData:
		return []byte{Version, byte(PingData), byte(protoType), 0, 0, 0, 0, 0}
	case PongData:
		return []byte{Version, byte(PongData), byte(protoType), 0, 0, 0, 0, 0}
	}

	return nil
}

func IsHeaderInvalid(message []byte) bool {

	if len(message) < 8 {
		return false
	}

	// version
	if message[0] != Version {
		return false
	}

	// message type
	if message[1] != byte(TextData) && message[1] != byte(BinData) && message[1] != byte(PingData) && message[1] != byte(PongData) {
		return false
	}

	// proto type
	if message[2] != byte(Json) && message[2] != byte(ProtoBuf) && message[2] != byte(Text) {
		return false
	}

	return true
}

func GetLen(message []byte) int {
	if message[1] == byte(TextData) {
		return int(message[3]+(message[7]|message[6]<<7|message[5]<<14|message[4]<<21)) + 8
	} else {
		return int(message[3]+(message[7]|message[6]<<8|message[5]<<16|message[4]<<24)) + 8
	}
}

func UnPack(message []byte) (version byte, messageType int, protoType int, route []byte, body []byte) {

	if !IsHeaderInvalid(message) {
		return
	}

	if message[1] == byte(TextData) {
		if int(message[3]+(message[7]|message[6]<<7|message[5]<<14|message[4]<<21))+8 != len(message) {
			return
		}
	} else {
		if int(message[3]+(message[7]|message[6]<<8|message[5]<<16|message[4]<<24))+8 != len(message) {
			return
		}
	}

	return message[0], int(message[1]), int(message[2]), message[8 : 8+message[3]], message[8+message[3]:]
}

func packText(route []byte, body []byte, protoType int) []byte {

	var bl = len(body)

	// data struct
	var data []byte

	// 0 version
	data = append(data, Version)

	// 1 message type
	data = append(data, byte(TextData))

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
	data = append(data, Version)

	// 1 message type
	data = append(data, byte(BinData))

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

func ParseMessage(bts []byte) ([]byte, []byte) {

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
