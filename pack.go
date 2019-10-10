/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-10-09 16:02
**/

package lemo

// 0 version
// 1 message type
// 2 proto type
// 3 route len
// 4 body len
// 5 body len
// 6 body len
// 7 body len

const (
	Text     byte = 1
	Json     byte = 2
	ProtoBuf byte = 3

	TextData byte = 1
	BinData  byte = 2
)

func Pack(route []byte, body []byte, protoType byte, messageType byte) []byte {

	switch messageType {
	case TextData:
		return packText(route, body, protoType)
	case BinData:
		return packBin(route, body, protoType)
	}

	return nil
}

func UnPack(message []byte) (version byte, messageType byte, protoType byte, route []byte, body []byte) {

	var mLen = len(message)

	if mLen < 8 {
		return
	}

	// version
	if message[0] != 'V' {
		return
	}

	// message type
	if message[1] != TextData && message[1] != BinData {
		return
	}

	// proto type
	if message[2] != Json && message[2] != ProtoBuf && message[2] != Text {
		return
	}

	if message[1] == TextData {
		if int(message[3]+(message[7]|message[6]<<7|message[5]<<14|message[4]<<21)) != mLen-8 {
			return
		}
	} else {
		if int(message[3]+(message[7]|message[6]<<8|message[5]<<16|message[4]<<24)) != mLen-8 {
			return
		}
	}

	return message[0], message[1], message[2], message[8 : 8+message[3]], message[8+message[3]:]
}

func packText(route []byte, body []byte, protoType byte) []byte {

	var bl = len(body)

	// data struct
	var data []byte

	// 0 version
	data = append(data, 'V')

	// 1 message type
	data = append(data, TextData)

	// 2 proto type
	data = append(data, protoType)

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
func packBin(route []byte, body []byte, protoType byte) []byte {

	var bl = len(body)

	// data struct
	var data []byte

	// 0 version
	data = append(data, 'V')

	// 1 message type
	data = append(data, BinData)

	// 2 proto type
	data = append(data, protoType)

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
