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

// 0 13
// 1 10
// 2 message type
// 3 route len
// 4 body len
// 5 body len

const Text byte = 1
const Json byte = 2
const ProtoBuf byte = 3

func Pack(route []byte, body []byte, messageType byte) []byte {

	// tips
	var data = []byte{13, 10}

	// message type
	data = append(data, messageType)

	// route len
	data = append(data, byte(len(route)))

	// body len 4
	data = append(data, byte(len(body)>>8))

	// body len 5
	data = append(data, byte(len(body)&0x00ff))

	data = append(data, route...)

	data = append(data, body...)

	return data
}

func UnPack(message []byte) ([]byte, []byte) {

	var mLen = len(message)

	if mLen < 6 {
		return nil, nil
	}

	if message[0] != 13 || message[1] != 10 || (message[2] != Json && message[2] != ProtoBuf && message[2] != Text) {
		return nil, nil
	}

	if int(message[3])+int(uint16(message[5])|uint16(message[4])<<8) != mLen-6 {
		return nil, nil
	}

	return message[6 : 6+message[3]], message[6+message[3]:]
}
