package lemo

import (
	"strings"

	"github.com/Lemo-yxk/tire"
)

type WebSocketClientGroupFunction func()

type WebSocketClientFunction func(c *WebSocketClient, msg *Receive) func() *Error

type WebSocketClientBefore func(c *WebSocketClient, msg *Receive) (Context, func() *Error)

type WebSocketClientAfter func(c *WebSocketClient, msg *Receive) func() *Error

var globalWebSocketClientPath string
var globalWebSocketClientBefore []WebSocketClientBefore
var globalWebSocketClientAfter []WebSocketClientAfter

func (client *WebSocketClient) FormatPath(path string) string {

	if client.IgnoreCase {
		path = strings.ToLower(path)
	}

	return path
}

func (client *WebSocketClient) Group(path string, v ...interface{}) {

	if v == nil {
		panic("Group function length is 0")
	}

	var g WebSocketClientGroupFunction

	for _, fn := range v {
		switch fn.(type) {
		case func():
			g = fn.(func())
		case []WebSocketClientBefore:
			globalWebSocketClientBefore = fn.([]WebSocketClientBefore)
		case []WebSocketClientAfter:
			globalWebSocketClientAfter = fn.([]WebSocketClientAfter)
		}
	}

	if g == nil {
		panic("Group function is nil")
	}

	globalWebSocketClientPath = path
	g()
	globalWebSocketClientPath = ""
	globalWebSocketClientBefore = nil
	globalWebSocketClientAfter = nil
}

func (client *WebSocketClient) SetRouter(path string, v ...interface{}) {

	path = client.FormatPath(globalWebSocketClientPath + path)

	if client.Router == nil {
		client.Router = new(tire.Tire)
	}

	var webSocketClientFunction WebSocketClientFunction
	var before []WebSocketClientBefore
	var after []WebSocketClientAfter

	var passBefore = false
	var passAfter = false
	var forceBefore = false
	var forceAfter = false

	for _, mark := range v {
		switch mark.(type) {
		case uint8:
			if mark.(uint8) == PassBefore {
				passBefore = true
			}
			if mark.(uint8) == PassAfter {
				passAfter = true
			}
			if mark.(uint8) == ForceBefore {
				forceBefore = true
			}
			if mark.(uint8) == ForceAfter {
				forceAfter = true
			}
		}
	}

	for _, fn := range v {
		switch fn.(type) {
		case func(c *WebSocketClient, msg *Receive) func() *Error:
			webSocketClientFunction = fn.(func(c *WebSocketClient, msg *Receive) func() *Error)
		case []WebSocketClientBefore:
			before = fn.([]WebSocketClientBefore)
		case []WebSocketClientAfter:
			after = fn.([]WebSocketClientAfter)
		}
	}

	if webSocketClientFunction == nil {
		println(path, "WebSocketServer function is nil")
		return
	}

	var cba = &CBA{}

	cba.WebSocketClientFunction = webSocketClientFunction

	cba.Before = append(globalWebSocketClientBefore, before...)
	if passBefore {
		cba.Before = nil
	}
	if forceBefore {
		cba.Before = before
	}

	cba.After = append(globalWebSocketClientAfter, after...)
	if passAfter {
		cba.After = nil
	}
	if forceAfter {
		cba.After = after
	}

	cba.Route = []byte(path)

	client.Router.Insert(path, cba)
}

func (client *WebSocketClient) GetRoute(path string) *tire.Tire {

	path = client.FormatPath(path)

	var pathB = []byte(path)

	if client.Router == nil {
		return nil
	}

	var t = client.Router.GetValue(pathB)

	if t == nil {
		return nil
	}

	var cba = t.Data.(*CBA)

	cba.Path = pathB

	return t
}

func (client *WebSocketClient) router(c *WebSocketClient, msg *ReceivePackage) {

	node := client.GetRoute(string(msg.Event))
	if node == nil {
		return
	}

	var cba = node.Data.(*CBA)

	var params = new(Params)
	params.Keys = node.Keys
	params.Values = node.ParseParams(cba.Path)

	var receive = &Receive{}
	receive.Message = msg
	receive.Context = nil
	receive.Params = params

	for _, before := range cba.Before {
		context, err := before(c, receive)
		if err != nil {
			if client.OnError != nil {
				client.OnError(err)
			}
			return
		}
		receive.Context = context
	}

	err := cba.WebSocketClientFunction(c, receive)
	if err != nil {
		if client.OnError != nil {
			client.OnError(err)
		}
		return
	}

	for _, after := range cba.After {
		err := after(c, receive)
		if err != nil {
			if client.OnError != nil {
				client.OnError(err)
			}
			return
		}
	}

}

type CBA struct {
	Path                    []byte
	Route                   []byte
	WebSocketClientFunction WebSocketClientFunction
	Before                  []WebSocketClientBefore
	After                   []WebSocketClientAfter
}
