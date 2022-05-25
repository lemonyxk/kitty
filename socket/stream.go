/**
* @program: lemon
*
* @description:
*
* @author: lemon
*
* @create: 2019-11-19 20:56
**/

package socket

import (
	"github.com/lemonyxk/kitty/v2/kitty"
)

type Stream[T any] struct {
	Data  []byte
	Event string

	Conn T

	Context kitty.Context
	Params  kitty.Params
	Logger  kitty.Logger
}
