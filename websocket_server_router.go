package lemo

import (
	"github.com/Lemo-yxk/lemo/exception"
	"runtime"
	"strconv"
	"strings"

	"github.com/Lemo-yxk/tire"
)

type WebSocketServerGroupFunction func(this *WebSocketServer)

type WebSocketServerFunction func(conn *WebSocket, receive *Receive) func() *exception.Error

type WebSocketServerBefore func(conn *WebSocket, receive *Receive) (Context, func() *exception.Error)

type WebSocketServerAfter func(conn *WebSocket, receive *Receive) func() *exception.Error

var webSocketServerGlobalBefore []WebSocketServerBefore
var webSocketServerGlobalAfter []WebSocketServerAfter

func SetWebSocketServerBefore(before ...WebSocketServerBefore) {
	webSocketServerGlobalBefore = append(webSocketServerGlobalBefore, before...)
}

func SetWebSocketServerAfter(after ...WebSocketServerAfter) {
	webSocketServerGlobalAfter = append(webSocketServerGlobalAfter, after...)
}

type webSocketServerGroup struct {
	path   string
	before []WebSocketServerBefore
	after  []WebSocketServerAfter
	socket *WebSocketServer
}

func (group *webSocketServerGroup) Route(path string) *webSocketServerGroup {
	group.path = path
	return group
}

func (group *webSocketServerGroup) Before(before ...WebSocketServerBefore) *webSocketServerGroup {
	group.before = append(group.before, before...)
	return group
}

func (group *webSocketServerGroup) After(after ...WebSocketServerAfter) *webSocketServerGroup {
	group.after = append(group.after, after...)
	return group
}

func (group *webSocketServerGroup) Handler(fn WebSocketServerGroupFunction) {
	fn(group.socket)
	group.socket.group = nil
}

type webSocketServerRoute struct {
	path        string
	before      []WebSocketServerBefore
	after       []WebSocketServerAfter
	socket      *WebSocketServer
	passBefore  bool
	forceBefore bool
	passAfter   bool
	forceAfter  bool
}

func (route *webSocketServerRoute) Route(path string) *webSocketServerRoute {
	route.path = path
	return route
}

func (route *webSocketServerRoute) Before(before ...WebSocketServerBefore) *webSocketServerRoute {
	route.before = append(route.before, before...)
	return route
}

func (route *webSocketServerRoute) PassBefore() *webSocketServerRoute {
	route.passBefore = true
	return route
}

func (route *webSocketServerRoute) ForceBefore() *webSocketServerRoute {
	route.forceBefore = true
	return route
}

func (route *webSocketServerRoute) After(after ...WebSocketServerAfter) *webSocketServerRoute {
	route.after = append(route.after, after...)
	return route
}

func (route *webSocketServerRoute) PassAfter() *webSocketServerRoute {
	route.passAfter = true
	return route
}

func (route *webSocketServerRoute) ForceAfter() *webSocketServerRoute {
	route.forceAfter = true
	return route
}

func (route *webSocketServerRoute) Handler(fn WebSocketServerFunction) {

	_, file, line, _ := runtime.Caller(1)

	var socket = route.socket
	var group = socket.group

	if group == nil {
		group = new(webSocketServerGroup)
	}

	var path = socket.formatPath(group.path + route.path)

	if socket.tire == nil {
		socket.tire = new(tire.Tire)
	}

	var wba = &WebSocketServerNode{}

	wba.Info = file + ":" + strconv.Itoa(line)

	wba.WebSocketServerFunction = fn

	wba.Before = append(group.before, route.before...)
	if route.passBefore {
		wba.Before = nil
	}
	if route.forceBefore {
		wba.Before = route.before
	}

	wba.After = append(group.after, route.after...)
	if route.passAfter {
		wba.After = nil
	}
	if route.forceAfter {
		wba.After = route.after
	}

	wba.Before = append(wba.Before, webSocketServerGlobalBefore...)
	wba.After = append(wba.After, webSocketServerGlobalAfter...)

	wba.Route = []byte(path)

	socket.tire.Insert(path, wba)

	route.socket.route = nil
}

func (socket *WebSocketServer) Group(path string) *webSocketServerGroup {

	var group = new(webSocketServerGroup)

	group.Route(path)

	group.socket = socket

	socket.group = group

	return group
}

func (socket *WebSocketServer) Route(path string) *webSocketServerRoute {

	var route = new(webSocketServerRoute)

	route.Route(path)

	route.socket = socket

	socket.route = route

	return route
}

func (socket *WebSocketServer) getRoute(path string) (*tire.Tire, []byte) {

	if socket.tire == nil {
		return nil, nil
	}

	path = socket.formatPath(path)

	var pathB = []byte(path)

	var t = socket.tire.GetValue(pathB)

	if t == nil {
		return nil, nil
	}

	return t, pathB
}

func (socket *WebSocketServer) router(conn *WebSocket, msg *ReceivePackage) {

	var node, formatPath = socket.getRoute(msg.Event)
	if node == nil {
		return
	}

	var nodeData = node.Data.(*WebSocketServerNode)

	var params = new(Params)
	params.Keys = node.Keys
	params.Values = node.ParseParams(formatPath)

	var receive = &Receive{}
	receive.Message = msg
	receive.Context = nil
	receive.Params = params

	for i := 0; i < len(nodeData.Before); i++ {
		context, err := nodeData.Before[i](conn, receive)
		if err != nil {
			if socket.OnError != nil {
				socket.OnError(err)
			}
			return
		}
		receive.Context = context
	}

	err := nodeData.WebSocketServerFunction(conn, receive)
	if err != nil {
		if socket.OnError != nil {
			socket.OnError(err)
		}
		return
	}

	for i := 0; i < len(nodeData.After); i++ {
		err := nodeData.After[i](conn, receive)
		if err != nil {
			if socket.OnError != nil {
				socket.OnError(err)
			}
			return
		}
	}

}

func (socket *WebSocketServer) formatPath(path string) string {
	if socket.IgnoreCase {
		path = strings.ToLower(path)
	}
	return path
}

type WebSocketServerNode struct {
	Info                    string
	Route                   []byte
	WebSocketServerFunction WebSocketServerFunction
	Before                  []WebSocketServerBefore
	After                   []WebSocketServerAfter
}
