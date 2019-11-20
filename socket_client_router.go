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
	"github.com/Lemo-yxk/lemo/exception"
	"runtime"
	"strconv"
	"strings"

	"github.com/Lemo-yxk/tire"
)

type SocketClientGroupFunction func(this *SocketClient)

type SocketClientFunction func(c *SocketClient, receive *Receive) func() *exception.Error

type SocketClientBefore func(c *SocketClient, receive *Receive) (Context, func() *exception.Error)

type SocketClientAfter func(c *SocketClient, receive *Receive) func() *exception.Error

var socketClientGlobalBefore []SocketClientBefore
var socketClientGlobalAfter []SocketClientAfter

func SetSocketClientBefore(before ...SocketClientBefore) {
	socketClientGlobalBefore = append(socketClientGlobalBefore, before...)
}

func SetSocketClientAfter(after ...SocketClientAfter) {
	socketClientGlobalAfter = append(socketClientGlobalAfter, after...)
}

type socketClientGroup struct {
	path   string
	before []SocketClientBefore
	after  []SocketClientAfter
	socket *SocketClient
}

func (group *socketClientGroup) Route(path string) *socketClientGroup {
	group.path = path
	return group
}

func (group *socketClientGroup) Before(before ...SocketClientBefore) *socketClientGroup {
	group.before = append(group.before, before...)
	return group
}

func (group *socketClientGroup) After(after ...SocketClientAfter) *socketClientGroup {
	group.after = append(group.after, after...)
	return group
}

func (group *socketClientGroup) Handler(fn SocketClientGroupFunction) {
	fn(group.socket)
	group.socket.group = nil
}

type socketClientRoute struct {
	path        string
	before      []SocketClientBefore
	after       []SocketClientAfter
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

func (route *socketClientRoute) Before(before ...SocketClientBefore) *socketClientRoute {
	route.before = append(route.before, before...)
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

func (route *socketClientRoute) After(after ...SocketClientAfter) *socketClientRoute {
	route.after = append(route.after, after...)
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

func (route *socketClientRoute) Handler(fn SocketClientFunction) {

	_, file, line, _ := runtime.Caller(1)

	var socket = route.socket
	var group = socket.group

	if group == nil {
		group = new(socketClientGroup)
	}

	var path = socket.formatPath(group.path + route.path)

	if socket.tire == nil {
		socket.tire = new(tire.Tire)
	}

	var cba = &SocketClientNode{}

	cba.Info = file + ":" + strconv.Itoa(line)

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

	cba.Before = append(cba.Before, socketClientGlobalBefore...)
	cba.After = append(cba.After, socketClientGlobalAfter...)

	cba.Route = []byte(path)

	socket.tire.Insert(path, cba)

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

func (client *SocketClient) getRoute(path string) (*tire.Tire, []byte) {

	if client.tire == nil {
		return nil, nil
	}

	path = client.formatPath(path)

	var pathB = []byte(path)

	var t = client.tire.GetValue(pathB)

	if t == nil {
		return nil, nil
	}

	return t, pathB
}

func (client *SocketClient) router(conn *SocketClient, msg *ReceivePackage) {

	var node, formatPath = client.getRoute(msg.Event)
	if node == nil {
		return
	}

	var nodeData = node.Data.(*SocketClientNode)

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

func (client *SocketClient) formatPath(path string) string {
	if client.IgnoreCase {
		path = strings.ToLower(path)
	}
	return path
}

type SocketClientNode struct {
	Info                 string
	Route                []byte
	SocketClientFunction SocketClientFunction
	Before               []SocketClientBefore
	After                []SocketClientAfter
}
