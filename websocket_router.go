package lemo

import (
	"strings"

	"github.com/Lemo-yxk/tire"
)

type WebSocketServerGroupFunction func()

type WebSocketServerFunction func(conn *WebSocket, msg *Receive) func() *Error

type WebSocketServerBefore func(conn *WebSocket, msg *Receive) (Context, func() *Error)

type WebSocketServerAfter func(conn *WebSocket, msg *Receive) func() *Error

var globalWebSocketServerPath string
var globalWebSocketServerBefore []WebSocketServerBefore
var globalWebSocketServerAfter []WebSocketServerAfter

func (socket *WebSocketServer) FormatPath(path string) string {

	if socket.IgnoreCase {
		path = strings.ToLower(path)
	}

	return path
}

func (socket *WebSocketServer) Group(path string, v ...interface{}) {

	if v == nil {
		panic("Group function length is 0")
	}

	var g WebSocketServerGroupFunction

	for _, fn := range v {
		switch fn.(type) {
		case func():
			g = fn.(func())
		case []WebSocketServerBefore:
			globalWebSocketServerBefore = fn.([]WebSocketServerBefore)
		case []WebSocketServerAfter:
			globalWebSocketServerAfter = fn.([]WebSocketServerAfter)
		}
	}

	if g == nil {
		panic("Group function is nil")
	}

	globalWebSocketServerPath = path
	g()
	globalWebSocketServerPath = ""
	globalWebSocketServerBefore = nil
	globalWebSocketServerAfter = nil
}

func (socket *WebSocketServer) SetRouter(path string, v ...interface{}) {

	path = socket.FormatPath(globalWebSocketServerPath + path)

	if socket.Router == nil {
		socket.Router = new(tire.Tire)
	}

	var webSocketServerFunction WebSocketServerFunction
	var before []WebSocketServerBefore
	var after []WebSocketServerAfter

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
		case func(conn *WebSocket, msg *Receive) func() *Error:
			webSocketServerFunction = fn.(func(conn *WebSocket, msg *Receive) func() *Error)
		case []WebSocketServerBefore:
			before = fn.([]WebSocketServerBefore)
		case []WebSocketServerAfter:
			after = fn.([]WebSocketServerAfter)
		}
	}

	if webSocketServerFunction == nil {
		println(path, "WebSocketServer function is nil")
		return
	}

	var wba = &WBA{}

	wba.WebSocketServerFunction = webSocketServerFunction

	wba.Before = append(globalWebSocketServerBefore, before...)
	if passBefore {
		wba.Before = nil
	}
	if forceBefore {
		wba.Before = before
	}

	wba.After = append(globalWebSocketServerAfter, after...)
	if passAfter {
		wba.After = nil
	}
	if forceAfter {
		wba.After = after
	}

	wba.Route = []byte(path)

	socket.Router.Insert(path, wba)
}

func (socket *WebSocketServer) GetRoute(path string) *tire.Tire {

	path = socket.FormatPath(path)

	var pathB = []byte(path)

	if socket.Router == nil {
		return nil
	}

	var t = socket.Router.GetValue(pathB)

	if t == nil {
		return nil
	}

	var wba = t.Data.(*WBA)

	wba.Path = pathB

	return t
}

func (socket *WebSocketServer) router(conn *WebSocket, msg *ReceivePackage) {

	node := socket.GetRoute(msg.Event)
	if node == nil {
		return
	}

	var wba = node.Data.(*WBA)

	var params = new(Params)
	params.Keys = node.Keys
	params.Values = node.ParseParams(wba.Path)

	var receive = &Receive{}
	receive.Message = msg
	receive.Context = nil
	receive.Params = params

	for _, before := range wba.Before {
		context, err := before(conn, receive)
		if err != nil {
			if socket.OnError != nil {
				socket.OnError(err)
			}
			return
		}
		receive.Context = context
	}

	err := wba.WebSocketServerFunction(conn, receive)
	if err != nil {
		if socket.OnError != nil {
			socket.OnError(err)
		}
		return
	}

	for _, after := range wba.After {
		err := after(conn, receive)
		if err != nil {
			if socket.OnError != nil {
				socket.OnError(err)
			}
			return
		}
	}

}

type WBA struct {
	Path                    []byte
	Route                   []byte
	WebSocketServerFunction WebSocketServerFunction
	Before                  []WebSocketServerBefore
	After                   []WebSocketServerAfter
}
