/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2022-05-22 04:20
**/

package router

type Node[T any, P any] struct {
	Data     P
	Info     string
	Desc     []string
	Route    []byte
	Function Func[T]
	Before   []Before[T]
	After    []After[T]
	Method   []string
}
