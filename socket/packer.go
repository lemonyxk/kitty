/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2023-04-22 17:49
**/

package socket

type Packer interface {
	Pack(async byte, messageType byte, code uint32, messageID uint64, route []byte, body []byte) error
	Push(msg []byte) error
}
