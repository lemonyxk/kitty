package ws

import (
	"github.com/tidwall/gjson"
)

func (socket *Socket) InitRouter() {
	socket.WebSocketRouter = make(map[string]WebSocketFunction)
}

func (socket *Socket) WSetRoute(route string, f WebSocketFunction) {
	socket.WebSocketRouter[route] = f
}

func (socket *Socket) WGetRouter(route string) WebSocketFunction {
	if f, ok := socket.WebSocketRouter[route]; ok {
		return f
	}
	return nil
}

func (socket *Socket) router(conn *Connection, message *Message) {

	// Json router
	if socket.TsProto == "json" {
		socket.jsonRouter(conn, message)
		return
	}

	// ProtoBuf router
	if socket.TsProto == "protobuf" {
		socket.protoBufRouter(conn, message)
		return
	}
}

func (socket *Socket) jsonRouter(conn *Connection, message *Message) {

	var event = gjson.GetBytes(message.Message, "Event").Str

	var f = socket.WGetRouter(event)

	if f == nil {
		return
	}

	f(conn, message, nil)
}

func (socket *Socket) protoBufRouter(conn *Connection, message *Message) {

}
