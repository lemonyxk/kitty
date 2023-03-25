/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2022-05-25 06:21
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
	Version byte = 'V'

	Unknown  byte = 0
	Text     byte = 1
	Bin      byte = 2
	Json     int  = 3
	ProtoBuf int  = 4

	Ping byte = 9
	Pong byte = 10
)

var PingMessage = []byte{0x0, 0x0, 0x9, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
var PongMessage = []byte{0x0, 0x0, 0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}

type Protocol interface {
	Decode(message []byte) (messageType byte, id int64, route []byte, body []byte)
	Encode(messageType byte, id int64, route []byte, body []byte) []byte
	Reader() func(n int, buf []byte, fn func(bytes []byte)) error
	HeadLen() int
	PackPing() []byte
	PackPong() []byte
	IsPong(messageType byte) bool
	IsPing(messageType byte) bool
	IsUnknown(messageType byte) bool
}
