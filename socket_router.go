/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-10-12 15:52
**/

package lemo

import (
	"strings"

	"github.com/Lemo-yxk/tire"
)

type SocketServerGroupFunction func(this *SocketServer)

type SocketServerFunction func(conn *Socket, receive *Receive) func() *Error

type SocketServerBefore func(conn *Socket, receive *Receive) (Context, func() *Error)

type SocketServerAfter func(conn *Socket, receive *Receive) func() *Error

type SocketServerGroup struct {
	path   string
	before []SocketServerBefore
	after  []SocketServerAfter
	socket *SocketServer
}

func (group *SocketServerGroup) Route(path string) *SocketServerGroup {
	group.path = path
	return group
}

func (group *SocketServerGroup) Before(before []SocketServerBefore) *SocketServerGroup {
	group.before = before
	return group
}

func (group *SocketServerGroup) After(after []SocketServerAfter) *SocketServerGroup {
	group.after = after
	return group
}

func (group *SocketServerGroup) Handler(fn SocketServerGroupFunction) {
	fn(group.socket)
	group.socket.group = nil
}

type SocketServerRoute struct {
	path        string
	before      []SocketServerBefore
	after       []SocketServerAfter
	socket      *SocketServer
	passBefore  bool
	forceBefore bool
	passAfter   bool
	forceAfter  bool
}

func (route *SocketServerRoute) Route(path string) *SocketServerRoute {
	route.path = path
	return route
}

func (route *SocketServerRoute) Before(before []SocketServerBefore) *SocketServerRoute {
	route.before = before
	return route
}

func (route *SocketServerRoute) PassBefore() *SocketServerRoute {
	route.passBefore = true
	return route
}

func (route *SocketServerRoute) ForceBefore() *SocketServerRoute {
	route.forceBefore = true
	return route
}

func (route *SocketServerRoute) After(after []SocketServerAfter) *SocketServerRoute {
	route.after = after
	return route
}

func (route *SocketServerRoute) PassAfter() *SocketServerRoute {
	route.passAfter = true
	return route
}

func (route *SocketServerRoute) ForceAfter() *SocketServerRoute {
	route.forceAfter = true
	return route
}

func (route *SocketServerRoute) Handler(fn SocketServerFunction) {

	var socket = route.socket
	var group = socket.group

	if group == nil {
		group = new(SocketServerGroup)
	}

	var path = socket.FormatPath(group.path + route.path)

	if socket.Router == nil {
		socket.Router = new(tire.Tire)
	}

	var sba = &SocketServerNode{}

	sba.SocketServerFunction = fn

	sba.Before = append(group.before, route.before...)
	if route.passBefore {
		sba.Before = nil
	}
	if route.forceBefore {
		sba.Before = route.before
	}

	sba.After = append(group.after, route.after...)
	if route.passAfter {
		sba.After = nil
	}
	if route.forceAfter {
		sba.After = route.after
	}

	sba.Route = []byte(path)

	socket.Router.Insert(path, sba)

	route.socket.route = nil
}

func (socket *SocketServer) Group(path string) *SocketServerGroup {

	var group = new(SocketServerGroup)

	group.Route(path)

	group.socket = socket

	socket.group = group

	return group
}

func (socket *SocketServer) Route(path string) *SocketServerRoute {

	var route = new(SocketServerRoute)

	route.Route(path)

	route.socket = socket

	socket.route = route

	return route
}

func (socket *SocketServer) getRoute(path string) *tire.Tire {

	path = socket.FormatPath(path)

	var pathB = []byte(path)

	if socket.Router == nil {
		return nil
	}

	var t = socket.Router.GetValue(pathB)

	if t == nil {
		return nil
	}

	var nodeData = t.Data.(*SocketServerNode)

	nodeData.Path = pathB

	return t
}

func (socket *SocketServer) router(conn *Socket, msg *ReceivePackage) {

	node := socket.getRoute(string(msg.Event))
	if node == nil {
		return
	}

	var nodeData = node.Data.(*SocketServerNode)

	var params = new(Params)
	params.Keys = node.Keys
	params.Values = node.ParseParams(nodeData.Path)

	var receive = &Receive{}
	receive.Message = msg
	receive.Context = nil
	receive.Params = params

	for _, before := range nodeData.Before {
		context, err := before(conn, receive)
		if err != nil {
			if socket.OnError != nil {
				socket.OnError(err)
			}
			return
		}
		receive.Context = context
	}

	err := nodeData.SocketServerFunction(conn, receive)
	if err != nil {
		if socket.OnError != nil {
			socket.OnError(err)
		}
		return
	}

	for _, after := range nodeData.After {
		err := after(conn, receive)
		if err != nil {
			if socket.OnError != nil {
				socket.OnError(err)
			}
			return
		}
	}

}

func (socket *SocketServer) FormatPath(path string) string {
	if socket.IgnoreCase {
		path = strings.ToLower(path)
	}
	return path
}

type SocketServerNode struct {
	Path                 []byte
	Route                []byte
	SocketServerFunction SocketServerFunction
	Before               []SocketServerBefore
	After                []SocketServerAfter
}
