package lemo

import (
	"strings"

	"github.com/Lemo-yxk/tire"
)

type WebSocketClientGroupFunction func(this *WebSocketClient)

type WebSocketClientFunction func(c *WebSocketClient, receive *Receive) func() *Error

type WebSocketClientBefore func(c *WebSocketClient, receive *Receive) (Context, func() *Error)

type WebSocketClientAfter func(c *WebSocketClient, receive *Receive) func() *Error

type WebSocketClientGroup struct {
	path   string
	before []WebSocketClientBefore
	after  []WebSocketClientAfter
	socket *WebSocketClient
}

func (group *WebSocketClientGroup) Route(path string) *WebSocketClientGroup {
	group.path = path
	return group
}

func (group *WebSocketClientGroup) Before(before []WebSocketClientBefore) *WebSocketClientGroup {
	group.before = before
	return group
}

func (group *WebSocketClientGroup) After(after []WebSocketClientAfter) *WebSocketClientGroup {
	group.after = after
	return group
}

func (group *WebSocketClientGroup) Handler(fn WebSocketClientGroupFunction) {
	fn(group.socket)
	group.socket.group = nil
}

type WebSocketClientRoute struct {
	path        string
	before      []WebSocketClientBefore
	after       []WebSocketClientAfter
	socket      *WebSocketClient
	passBefore  bool
	forceBefore bool
	passAfter   bool
	forceAfter  bool
}

func (route *WebSocketClientRoute) Route(path string) *WebSocketClientRoute {
	route.path = path
	return route
}

func (route *WebSocketClientRoute) Before(before []WebSocketClientBefore) *WebSocketClientRoute {
	route.before = before
	return route
}

func (route *WebSocketClientRoute) PassBefore() *WebSocketClientRoute {
	route.passBefore = true
	return route
}

func (route *WebSocketClientRoute) ForceBefore() *WebSocketClientRoute {
	route.forceBefore = true
	return route
}

func (route *WebSocketClientRoute) After(after []WebSocketClientAfter) *WebSocketClientRoute {
	route.after = after
	return route
}

func (route *WebSocketClientRoute) PassAfter() *WebSocketClientRoute {
	route.passAfter = true
	return route
}

func (route *WebSocketClientRoute) ForceAfter() *WebSocketClientRoute {
	route.forceAfter = true
	return route
}

func (route *WebSocketClientRoute) Handler(fn WebSocketClientFunction) {

	var socket = route.socket
	var group = socket.group

	if group == nil {
		group = new(WebSocketClientGroup)
	}

	var path = socket.FormatPath(group.path + route.path)

	if socket.Router == nil {
		socket.Router = new(tire.Tire)
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

	socket.Router.Insert(path, cba)

	route.socket.route = nil
}

func (client *WebSocketClient) Group(path string) *WebSocketClientGroup {

	var group = new(WebSocketClientGroup)

	group.Route(path)

	group.socket = client

	client.group = group

	return group
}

func (client *WebSocketClient) Route(path string) *WebSocketClientRoute {

	var route = new(WebSocketClientRoute)

	route.Route(path)

	route.socket = client

	client.route = route

	return route
}

func (client *WebSocketClient) getRoute(path string) *tire.Tire {

	path = client.FormatPath(path)

	var pathB = []byte(path)

	if client.Router == nil {
		return nil
	}

	var t = client.Router.GetValue(pathB)

	if t == nil {
		return nil
	}

	var nodeData = t.Data.(*WebSocketClientNode)

	nodeData.Path = pathB

	return t
}

func (client *WebSocketClient) router(conn *WebSocketClient, msg *ReceivePackage) {

	node := client.getRoute(string(msg.Event))
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

	err := nodeData.WebSocketClientFunction(conn, receive)
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

func (client *WebSocketClient) FormatPath(path string) string {
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
