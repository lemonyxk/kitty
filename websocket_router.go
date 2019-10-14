package lemo

import (
	"strings"

	"github.com/Lemo-yxk/tire"
)

type WebSocketServerGroupFunction func(socket *WebSocketServer)

type WebSocketServerFunction func(conn *WebSocket, receive *Receive) func() *Error

type WebSocketServerBefore func(conn *WebSocket, receive *Receive) (Context, func() *Error)

type WebSocketServerAfter func(conn *WebSocket, receive *Receive) func() *Error

type WebSocketServerGroup struct {
	path   string
	before []WebSocketServerBefore
	after  []WebSocketServerAfter
	socket *WebSocketServer
}

func (group *WebSocketServerGroup) Route(path string) *WebSocketServerGroup {
	group.path = path
	return group
}

func (group *WebSocketServerGroup) Before(before []WebSocketServerBefore) *WebSocketServerGroup {
	group.before = before
	return group
}

func (group *WebSocketServerGroup) After(after []WebSocketServerAfter) *WebSocketServerGroup {
	group.after = after
	return group
}

func (group *WebSocketServerGroup) Handler(fn WebSocketServerGroupFunction) {
	fn(group.socket)
	group.socket.group = nil
}

type WebSocketServerRoute struct {
	path        string
	before      []WebSocketServerBefore
	after       []WebSocketServerAfter
	socket      *WebSocketServer
	passBefore  bool
	forceBefore bool
	passAfter   bool
	forceAfter  bool
}

func (route *WebSocketServerRoute) Route(path string) *WebSocketServerRoute {
	route.path = path
	return route
}

func (route *WebSocketServerRoute) Before(before []WebSocketServerBefore) *WebSocketServerRoute {
	route.before = before
	return route
}

func (route *WebSocketServerRoute) PassBefore() *WebSocketServerRoute {
	route.passBefore = true
	return route
}

func (route *WebSocketServerRoute) ForceBefore() *WebSocketServerRoute {
	route.forceBefore = true
	return route
}

func (route *WebSocketServerRoute) After(after []WebSocketServerAfter) *WebSocketServerRoute {
	route.after = after
	return route
}

func (route *WebSocketServerRoute) PassAfter() *WebSocketServerRoute {
	route.passAfter = true
	return route
}

func (route *WebSocketServerRoute) ForceAfter() *WebSocketServerRoute {
	route.forceAfter = true
	return route
}

func (route *WebSocketServerRoute) Handler(fn WebSocketServerFunction) {

	var socket = route.socket
	var group = socket.group

	if group == nil {
		group = new(WebSocketServerGroup)
	}

	var path = socket.FormatPath(group.path + route.path)

	if socket.Router == nil {
		socket.Router = new(tire.Tire)
	}

	var wba = &WBA{}

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

	wba.Route = []byte(path)

	socket.Router.Insert(path, wba)

	route.socket.route = nil
}

func (socket *WebSocketServer) Group(path string) *WebSocketServerGroup {

	var group = new(WebSocketServerGroup)

	group.Route(path)

	group.socket = socket

	socket.group = group

	return group
}

func (socket *WebSocketServer) Route(path string) *WebSocketServerRoute {

	var route = new(WebSocketServerRoute)

	route.Route(path)

	route.socket = socket

	socket.route = route

	return route
}

func (socket *WebSocketServer) getRoute(path string) *tire.Tire {

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

	node := socket.getRoute(string(msg.Event))
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

func (socket *WebSocketServer) FormatPath(path string) string {
	if socket.IgnoreCase {
		path = strings.ToLower(path)
	}
	return path
}

type WBA struct {
	Path                    []byte
	Route                   []byte
	WebSocketServerFunction WebSocketServerFunction
	Before                  []WebSocketServerBefore
	After                   []WebSocketServerAfter
}
