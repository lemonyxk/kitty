/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2022-05-22 19:06
**/

package router

type Func[T any] func(stream T) error

type Before[T any] func(stream T) error

type After[T any] func(stream T) error
