package lemo

import (
	"strings"

	"github.com/Lemo-yxk/tire"
)

type WebSocketGroupFunction func()

type WebSocketFunction func(conn *Connection, msg *Receive) func() *Error

type WebSocketBefore func(conn *Connection, msg *MessagePackage) (Context, func() *Error)

type WebSocketAfter func(conn *Connection, msg *MessagePackage) func() *Error

var globalSocketPath string
var globalSocketBefore []WebSocketBefore
var globalSocketAfter []WebSocketAfter

func (socket *Socket) FormatPath(path string) string {

	if socket.IgnoreCase {
		path = strings.ToLower(path)
	}

	return path
}

func (socket *Socket) Group(path string, v ...interface{}) {

	if v == nil {
		panic("Group function length is 0")
	}

	var g WebSocketGroupFunction

	for _, fn := range v {
		switch fn.(type) {
		case func():
			g = fn.(func())
		case []WebSocketBefore:
			globalSocketBefore = fn.([]WebSocketBefore)
		case []WebSocketAfter:
			globalSocketAfter = fn.([]WebSocketAfter)
		}
	}

	if g == nil {
		panic("Group function is nil")
	}

	globalSocketPath = path
	g()
	globalSocketPath = ""
	globalSocketBefore = nil
	globalSocketAfter = nil
}

func (socket *Socket) SetRouter(path string, v ...interface{}) {

	path = socket.FormatPath(globalSocketPath + path)

	if socket.Router == nil {
		socket.Router = new(tire.Tire)
	}

	var webSocketFunction WebSocketFunction
	var before []WebSocketBefore
	var after []WebSocketAfter

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
		case func(conn *Connection, msg *Receive) func() *Error:
			webSocketFunction = fn.(func(conn *Connection, msg *Receive) func() *Error)
		case []WebSocketBefore:
			before = fn.([]WebSocketBefore)
		case []WebSocketAfter:
			after = fn.([]WebSocketAfter)
		}
	}

	if webSocketFunction == nil {
		println(path, "WebSocket function is nil")
		return
	}

	var wba = &WBA{}

	wba.WebSocketFunction = webSocketFunction

	wba.Before = append(globalSocketBefore, before...)
	if passBefore {
		wba.Before = nil
	}
	if forceBefore {
		wba.Before = before
	}

	wba.After = append(globalSocketAfter, after...)
	if passAfter {
		wba.After = nil
	}
	if forceAfter {
		wba.After = after
	}

	wba.Route = []byte(path)

	socket.Router.Insert(path, wba)
}

func (socket *Socket) GetRoute(path string) *tire.Tire {

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

func (socket *Socket) router(conn *Connection, msg *MessagePackage) {

	switch socket.TransportType {
	case Json:
		socket.jsonRouter(conn, msg)
	case ProtoBuf:
		socket.protoBufRouter(conn, msg)
	}

}

func (socket *Socket) jsonRouter(conn *Connection, msg *MessagePackage) {

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
		context, err := before(conn, msg)
		if err != nil {
			if socket.OnError != nil {
				socket.OnError(err)
			}
			return
		}
		receive.Context = context
	}

	err := wba.WebSocketFunction(conn, receive)
	if err != nil {
		if socket.OnError != nil {
			socket.OnError(err)
		}
		return
	}

	for _, after := range wba.After {
		err := after(conn, msg)
		if err != nil {
			if socket.OnError != nil {
				socket.OnError(err)
			}
			return
		}
	}
}

func (socket *Socket) protoBufRouter(conn *Connection, msg *MessagePackage) {

}

type WBA struct {
	Path              []byte
	Route             []byte
	WebSocketFunction WebSocketFunction
	Before            []WebSocketBefore
	After             []WebSocketAfter
}
