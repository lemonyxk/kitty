/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2022-05-24 06:14
**/

package errors

var (
	ConnNotFount     = New("conn not found")
	RouteNotFount    = New("route not found")
	MethodNotAllowed = New("method not allowed")
	ClientClosed     = New("client closed")
	NilError         = New("nil error")
	ServerClosed     = New("server closed")
	Timeout          = New("timeout")
	Invalid          = New("invalid")
	MaximumExceeded  = New("maximum exceeded")
	AssertionFailed  = New("assertion failed")
	StopPropagation  = New("stop propagation")
)
