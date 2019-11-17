package lemo

import (
	"strings"

	"github.com/Lemo-yxk/tire"
)

type WebSocketClientGroupFunction func(this *WebSocketClient)

type WebSocketClientFunction func(c *WebSocketClient, receive *Receive) func() *Error

type WebSocketClientBefore func(c *WebSocketClient, receive *Receive) (Context, func() *Error)

type WebSocketClientAfter func(c *WebSocketClient, receive *Receive) func() *Error

type webSocketClientGroup struct {
	path   string
	before []WebSocketClientBefore
	after  []WebSocketClientAfter
	socket *WebSocketClient
}

func (group *webSocketClientGroup) Route(path string) *webSocketClientGroup {
	group.path = path
	return group
}

func (group *webSocketClientGroup) Before(before ...WebSocketClientBefore) *webSocketClientGroup {
	group.before = append(group.before, before...)
	return group
}

func (group *webSocketClientGroup) After(after ...WebSocketClientAfter) *webSocketClientGroup {
	group.after = append(group.after, after...)
	return group
}

func (group *webSocketClientGroup) Handler(fn WebSocketClientGroupFunction) {
	fn(group.socket)
	group.socket.group = nil
}

type webSocketClientRoute struct {
	path        string
	before      []WebSocketClientBefore
	after       []WebSocketClientAfter
	socket      *WebSocketClient
	passBefore  bool
	forceBefore bool
	passAfter   bool
	forceAfter  bool
}

func (route *webSocketClientRoute) Route(path string) *webSocketClientRoute {
	route.path = path
	return route
}

func (route *webSocketClientRoute) Before(before ...WebSocketClientBefore) *webSocketClientRoute {
	route.before = append(route.before, before...)
	return route
}

func (route *webSocketClientRoute) PassBefore() *webSocketClientRoute {
	route.passBefore = true
	return route
}

func (route *webSocketClientRoute) ForceBefore() *webSocketClientRoute {
	route.forceBefore = true
	return route
}

func (route *webSocketClientRoute) After(after ...WebSocketClientAfter) *webSocketClientRoute {
	route.after = append(route.after, after...)
	return route
}

func (route *webSocketClientRoute) PassAfter() *webSocketClientRoute {
	route.passAfter = true
	return route
}

func (route *webSocketClientRoute) ForceAfter() *webSocketClientRoute {
	route.forceAfter = true
	return route
}

func (route *webSocketClientRoute) Handler(fn WebSocketClientFunction) {

	var socket = route.socket
	var group = socket.group

	if group == nil {
		group = new(webSocketClientGroup)
	}

	var path = socket.formatPath(group.path + route.path)

	if socket.tire == nil {
		socket.tire = new(tire.Tire)
	}

	var cba = &WebSocketClientNode{}

	cba.WebSocketClientFunction = fn

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

	socket.tire.Insert(path, cba)

	route.socket.route = nil
}

func (client *WebSocketClient) Group(path string) *webSocketClientGroup {

	var group = new(webSocketClientGroup)

	group.Route(path)

	group.socket = client

	client.group = group

	return group
}

func (client *WebSocketClient) Route(path string) *webSocketClientRoute {

	var route = new(webSocketClientRoute)

	route.Route(path)

	route.socket = client

	client.route = route

	return route
}

func (client *WebSocketClient) getRoute(path string) *tire.Tire {

	path = client.formatPath(path)

	var pathB = []byte(path)

	if client.tire == nil {
		return nil
	}

	var t = client.tire.GetValue(pathB)

	if t == nil {
		return nil
	}

	var nodeData = t.Data.(*WebSocketClientNode)

	nodeData.Path = pathB

	return t
}

func (client *WebSocketClient) router(conn *WebSocketClient, msg *ReceivePackage) {

	var node = client.getRoute(msg.Event)
	if node == nil {
		return
	}

	var nodeData = node.Data.(*WebSocketClientNode)

	var params = new(Params)
	params.Keys = node.Keys
	params.Values = node.ParseParams(nodeData.Path)

	var receive = &Receive{}
	receive.Message = msg
	receive.Context = nil
	receive.Params = params

	for i := 0; i < len(nodeData.Before); i++ {
		context, err := nodeData.Before[i](conn, receive)
		if err != nil {
			if client.OnError != nil {
				client.OnError(err)
			}
			return
		}
		receive.Context = context
	}

	err := nodeData.WebSocketClientFunction(conn, receive)
	if err != nil {
		if client.OnError != nil {
			client.OnError(err)
		}
		return
	}

	for i := 0; i < len(nodeData.After); i++ {
		err := nodeData.After[i](conn, receive)
		if err != nil {
			if client.OnError != nil {
				client.OnError(err)
			}
			return
		}
	}

}

func (client *WebSocketClient) formatPath(path string) string {
	if client.IgnoreCase {
		path = strings.ToLower(path)
	}
	return path
}

type WebSocketClientNode struct {
	Path                    []byte
	Route                   []byte
	WebSocketClientFunction WebSocketClientFunction
	Before                  []WebSocketClientBefore
	After                   []WebSocketClientAfter
}
