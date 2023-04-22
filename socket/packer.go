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
	Pack(messageType byte, code int, messageID int64, route []byte, body []byte) error
	Push(msg []byte) error
}