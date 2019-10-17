/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-10-16 16:10
**/

package lemo

import (
	"strings"

	"github.com/Lemo-yxk/tire"
)

type socketClientGroupFunction func(this *SocketClient)

type socketClientFunction func(c *SocketClient, receive *Receive) func() *Error

type socketClientBefore func(c *SocketClient, receive *Receive) (Context, func() *Error)

type socketClientAfter func(c *SocketClient, receive *Receive) func() *Error

type socketClientGroup struct {
	path   string
	before []socketClientBefore
	after  []socketClientAfter
	socket *SocketClient
}

func (group *socketClientGroup) Route(path string) *socketClientGroup {
	group.path = path
	return group
}

func (group *socketClientGroup) Before(before []socketClientBefore) *socketClientGroup {
	group.before = before
	return group
}

func (group *socketClientGroup) After(after []socketClientAfter) *socketClientGroup {
	group.after = after
	return group
}

func (group *socketClientGroup) Handler(fn socketClientGroupFunction) {
	fn(group.socket)
	group.socket.group = nil
}

type socketClientRoute struct {
	path        string
	before      []socketClientBefore
	after       []socketClientAfter
	socket      *SocketClient
	passBefore  bool
	forceBefore bool
	passAfter   bool
	forceAfter  bool
}

func (route *socketClientRoute) Route(path string) *socketClientRoute {
	route.path = path
	return route
}

func (route *socketClientRoute) Before(before []socketClientBefore) *socketClientRoute {
	route.before = before
	return route
}

func (route *socketClientRoute) PassBefore() *socketClientRoute {
	route.passBefore = true
	return route
}

func (route *socketClientRoute) ForceBefore() *socketClientRoute {
	route.forceBefore = true
	return route
}

func (route *socketClientRoute) After(after []socketClientAfter) *socketClientRoute {
	route.after = after
	return route
}

func (route *socketClientRoute) PassAfter() *socketClientRoute {
	route.passAfter = true
	return route
}

func (route *socketClientRoute) ForceAfter() *socketClientRoute {
	route.forceAfter = true
	return route
}

func (route *socketClientRoute) Handler(fn socketClientFunction) {

	var socket = route.socket
	var group = socket.group

	if group == nil {
		group = new(socketClientGroup)
	}

	var path = socket.formatPath(group.path + route.path)

	if socket.Router == nil {
		socket.Router = new(tire.Tire)
	}

	var cba = &SocketClientNode{}

	cba.SocketClientFunction = fn

	cba.Before = append(group.before, route.before...)
	if route.passBefore {
		cba.Before = nil
	}
	if route.forceBefore {
		cba.Before = route.before
	}

	cba.After = append(group.after, route.after...)
	if route.passAfter {
		cba.After = nil
	}
	if route.forceAfter {
		cba.After = route.after
	}

	cba.Route = []byte(path)

	socket.Router.Insert(path, cba)

	route.socket.route = nil
}

func (client *SocketClient) Group(path string) *socketClientGroup {

	var group = new(socketClientGroup)

	group.Route(path)

	group.socket = client

	client.group = group

	return group
}

func (client *SocketClient) Route(path string) *socketClientRoute {

	var route = new(socketClientRoute)

	route.Route(path)

	route.socket = client

	client.route = route

	return route
}

func (client *SocketClient) getRoute(path string) *tire.Tire {

	path = client.formatPath(path)

	var pathB = []byte(path)

	if client.Router == nil {
		return nil
	}

	var t = client.Router.GetValue(pathB)

	if t == nil {
		return nil
	}

	var nodeData = t.Data.(*SocketClientNode)

	nodeData.Path = pathB

	return t
}

func (client *SocketClient) router(conn *SocketClient, msg *ReceivePackage) {

	node := client.getRoute(string(msg.Event))
	if node == nil {
		return
	}

	var nodeData = node.Data.(*SocketClientNode)

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
			if client.OnError != nil {
				client.OnError(err)
			}
			return
		}
		receive.Context = context
	}

	err := nodeData.SocketClientFunction(conn, receive)
	if err != nil {
		if client.OnError != nil {
			client.OnError(err)
		}
		return
	}

	for _, after := range nodeData.After {
		err := after(conn, receive)
		if err != nil {
			if client.OnError != nil {
				client.OnError(err)
			}
			return
		}
	}

}

func (client *SocketClient) formatPath(path string) string {
	if client.IgnoreCase {
		path = strings.ToLower(path)
	}
	return path
}

type SocketClientNode struct {
	Path                 []byte
	Route                []byte
	SocketClientFunction socketClientFunction
	Before               []socketClientBefore
	After                []socketClientAfter
}
