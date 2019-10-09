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

// 0 message type
// 1 route len
// 2 body len
// 3 body len
// 4 body len
// 5 body len

const Text byte = 1
const Json byte = 2
const ProtoBuf byte = 3

func Pack(route []byte, body []byte, protoType byte, messageType int) []byte {

	switch messageType {
	case TextMessage:
		return packText(route, body, protoType)
	case BinaryMessage:
		return packBin(route, body, protoType)
	}

	return nil
}

func UnPack(message []byte, messageType int) ([]byte, []byte, byte) {

	if messageType != TextMessage && messageType != BinaryMessage {
		return nil, nil, 0
	}

	var mLen = len(message)

	if mLen < 6 {
		return nil, nil, 0
	}

	if message[0] != Json && message[0] != ProtoBuf && message[0] != Text {
		return nil, nil, 0
	}

	if messageType == TextMessage {
		if int(message[1]+(message[5]|message[4]<<7|message[3]<<14|message[2]<<21)) != mLen-6 {
			return nil, nil, 0
		}
	} else {
		if int(message[1]+(message[5]|message[4]<<8|message[3]<<16|message[2]<<24)) != mLen-6 {
			return nil, nil, 0
		}
	}

	return message[6 : 6+message[1]], message[6+message[1]:], message[0]
}

func packText(route []byte, body []byte, protoType byte) []byte {

	var bl = len(body)

	// tips
	var data []byte

	// 1 message type
	data = append(data, protoType)

	// 2 route len
	data = append(data, byte(len(route)&0x007f))

	// 3 body len
	data = append(data, byte(bl>>21&0x007f))

	// 4 body len
	data = append(data, byte(bl>>14&0x007f))

	// 5 body len
	data = append(data, byte(bl>>7&0x007f))

	// 6 body len
	data = append(data, byte(bl&0x007f))

	data = append(data, route...)

	data = append(data, body...)

	return data

}
func packBin(route []byte, body []byte, protoType byte) []byte {

	var bl = len(body)

	// tips
	var data []byte

	// 1 message type
	data = append(data, protoType)

	// 2 route len
	data = append(data, byte(len(route)&0x00ff))

	// 3 body len
	data = append(data, byte(bl>>28&0x00ff))

	// 4 body len
	data = append(data, byte(bl>>16&0x00ff))

	// 5 body len
	data = append(data, byte(bl>>8&0x00ff))

	// 6 body len
	data = append(data, byte(bl&0x00ff))

	data = append(data, route...)

	data = append(data, body...)

	return data
}
