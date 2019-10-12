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

type SocketServerGroupFunction func(socket *SocketServer)

type SocketServerFunction func(conn *Socket, msg *Receive) func() *Error

type SocketServerBefore func(conn *Socket, msg *Receive) (Context, func() *Error)

type SocketServerAfter func(conn *Socket, msg *Receive) func() *Error

func (socket *SocketServer) FormatPath(path string) string {
	if socket.IgnoreCase {
		path = strings.ToLower(path)
	}
	return path
}

type Group struct {
	path   string
	before []SocketServerBefore
	after  []SocketServerAfter
	socket *SocketServer
}

func (group *Group) Route(path string) *Group {
	group.path = path
	return group
}

func (group *Group) Before(before []SocketServerBefore) *Group {
	group.before = before
	return group
}

func (group *Group) After(after []SocketServerAfter) *Group {
	group.after = after
	return group
}

func (socket *SocketServer) Group(path string) *Group {

	var group = new(Group)

	group.Route(path)

	group.socket = socket

	socket.group = group

	return group
}

func (group *Group) Handler(fn SocketServerGroupFunction) {
	fn(group.socket)
	group.socket.group = nil
}

type Route struct {
	path        string
	before      []SocketServerBefore
	after       []SocketServerAfter
	socket      *SocketServer
	passBefore  bool
	forceBefore bool
	passAfter   bool
	forceAfter  bool
}

func (route *Route) Route(path string) *Route {
	route.path = path
	return route
}

func (route *Route) Before(before []SocketServerBefore) *Route {
	route.before = before
	return route
}

func (route *Route) PassBefore() *Route {
	route.passBefore = true
	return route
}

func (route *Route) ForceBefore() *Route {
	route.forceBefore = true
	return route
}

func (route *Route) After(after []SocketServerAfter) *Route {
	route.after = after
	return route
}

func (route *Route) PassAfter() *Route {
	route.passAfter = true
	return route
}

func (route *Route) ForceAfter() *Route {
	route.forceAfter = true
	return route
}

func (socket *SocketServer) Route(path string) *Route {

	var route = new(Route)

	route.Route(path)

	route.socket = socket

	socket.route = route

	return route
}

func (route *Route) Handler(fn SocketServerFunction) {

	var socket = route.socket
	var group = socket.group

	if group == nil {
		group = new(Group)
	}

	var path = socket.FormatPath(group.path + route.path)

	if socket.Router == nil {
		socket.Router = new(tire.Tire)
	}

	var sba = &SBA{}

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

	var sba = t.Data.(*SBA)

	sba.Path = pathB

	return t
}

func (socket *SocketServer) router(conn *Socket, msg *ReceivePackage) {

	node := socket.getRoute(string(msg.Event))
	if node == nil {
		return
	}

	var sba = node.Data.(*SBA)

	var params = new(Params)
	params.Keys = node.Keys
	params.Values = node.ParseParams(sba.Path)

	var receive = &Receive{}
	receive.Message = msg
	receive.Context = nil
	receive.Params = params

	for _, before := range sba.Before {
		context, err := before(conn, receive)
		if err != nil {
			if socket.OnError != nil {
				socket.OnError(err)
			}
			return
		}
		receive.Context = context
	}

	err := sba.SocketServerFunction(conn, receive)
	if err != nil {
		if socket.OnError != nil {
			socket.OnError(err)
		}
		return
	}

	for _, after := range sba.After {
		err := after(conn, receive)
		if err != nil {
			if socket.OnError != nil {
				socket.OnError(err)
			}
			return
		}
	}

}

type SBA struct {
	Path                 []byte
	Route                []byte
	SocketServerFunction SocketServerFunction
	Before               []SocketServerBefore
	After                []SocketServerAfter
}
