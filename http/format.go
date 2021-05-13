/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2021-05-13 19:32
**/

package http

type JsonFormat struct {
	Status string      `json:"status"`
	Code   int         `json:"code"`
	Msg    interface{} `json:"msg"`
}
