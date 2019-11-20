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
	"github.com/Lemo-yxk/lemo/exception"
	"runtime"
	"strconv"
	"strings"

	"github.com/Lemo-yxk/tire"
)

type SocketServerGroupFunction func(this *SocketServer)

type SocketServerFunction func(conn *Socket, receive *Receive) func() *exception.Error

type SocketServerBefore func(conn *Socket, receive *Receive) (Context, func() *exception.Error)

type SocketServerAfter func(conn *Socket, receive *Receive) func() *exception.Error

var socketServerGlobalBefore []SocketServerBefore
var socketServerGlobalAfter []SocketServerAfter

func SetSocketServerBefore(before ...SocketServerBefore) {
	socketServerGlobalBefore = append(socketServerGlobalBefore, before...)
}

func SetSocketServerAfter(after ...SocketServerAfter) {
	socketServerGlobalAfter = append(socketServerGlobalAfter, after...)
}

type socketServerGroup struct {
	path   string
	before []SocketServerBefore
	after  []SocketServerAfter
	socket *SocketServer
}

func (group *socketServerGroup) Route(path string) *socketServerGroup {
	group.path = path
	return group
}

func (group *socketServerGroup) Before(before ...SocketServerBefore) *socketServerGroup {
	group.before = append(group.before, before...)
	return group
}

func (group *socketServerGroup) After(after ...SocketServerAfter) *socketServerGroup {
	group.after = append(group.after, after...)
	return group
}

func (group *socketServerGroup) Handler(fn SocketServerGroupFunction) {
	fn(group.socket)
	group.socket.group = nil
}

type socketServerRoute struct {
	path        string
	before      []SocketServerBefore
	after       []SocketServerAfter
	socket      *SocketServer
	passBefore  bool
	forceBefore bool
	passAfter   bool
	forceAfter  bool
}

func (route *socketServerRoute) Route(path string) *socketServerRoute {
	route.path = path
	return route
}

func (route *socketServerRoute) Before(before ...SocketServerBefore) *socketServerRoute {
	route.before = append(route.before, before...)
	return route
}

func (route *socketServerRoute) PassBefore() *socketServerRoute {
	route.passBefore = true
	return route
}

func (route *socketServerRoute) ForceBefore() *socketServerRoute {
	route.forceBefore = true
	return route
}

func (route *socketServerRoute) After(after ...SocketServerAfter) *socketServerRoute {
	route.after = append(route.after, after...)
	return route
}

func (route *socketServerRoute) PassAfter() *socketServerRoute {
	route.passAfter = true
	return route
}

func (route *socketServerRoute) ForceAfter() *socketServerRoute {
	route.forceAfter = true
	return route
}

func (route *socketServerRoute) Handler(fn SocketServerFunction) {

	_, file, line, _ := runtime.Caller(1)

	var socket = route.socket
	var group = socket.group

	if group == nil {
		group = new(socketServerGroup)
	}

	var path = socket.formatPath(group.path + route.path)

	if socket.tire == nil {
		socket.tire = new(tire.Tire)
	}

	var sba = &SocketServerNode{}

	sba.Info = file + ":" + strconv.Itoa(line)

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

	sba.Before = append(sba.Before, socketServerGlobalBefore...)
	sba.After = append(sba.After, socketServerGlobalAfter...)

	sba.Route = []byte(path)

	socket.tire.Insert(path, sba)

	route.socket.route = nil
}

func (socket *SocketServer) Group(path string) *socketServerGroup {

	var group = new(socketServerGroup)

	group.Route(path)

	group.socket = socket

	socket.group = group

	return group
}

func (socket *SocketServer) Route(path string) *socketServerRoute {

	var route = new(socketServerRoute)

	route.Route(path)

	route.socket = socket

	socket.route = route

	return route
}

func (socket *SocketServer) getRoute(path string) (*tire.Tire, []byte) {

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

func (socket *SocketServer) router(conn *Socket, msg *ReceivePackage) {

	var node, formatPath = socket.getRoute(msg.Event)
	if node == nil {
		return
	}

	var nodeData = node.Data.(*SocketServerNode)

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

	err := nodeData.SocketServerFunction(conn, receive)
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

func (socket *SocketServer) formatPath(path string) string {
	if socket.IgnoreCase {
		path = strings.ToLower(path)
	}
	return path
}

type SocketServerNode struct {
	Info                 string
	Route                []byte
	SocketServerFunction SocketServerFunction
	Before               []SocketServerBefore
	After                []SocketServerAfter
}
