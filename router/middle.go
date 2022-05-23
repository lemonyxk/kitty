/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2022-05-22 19:06
**/

package router

type Middle[T any] func(T)

type Func[T any] func(stream T) error

type Before[T any] func(stream T) error

type After[T any] func(stream T) error
