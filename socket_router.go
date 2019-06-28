package ws

import (
	"github.com/tidwall/gjson"
)

func (socket *Socket) InitRouter() {
	socket.WebSocketRouter = make(map[string]WebSocketServerFunction)
}

func (socket *Socket) SetRouter(route string, f WebSocketServerFunction) {
	socket.WebSocketRouter[route] = f
}

func (socket *Socket) GetRouter(route string) WebSocketServerFunction {
	if f, ok := socket.WebSocketRouter[route]; ok {
		return f
	}
	return nil
}

func (socket *Socket) router(conn *Connection, message *Message) {

	switch socket.TsProto {
	case Json:
		socket.jsonRouter(conn, message)
	case ProtoBuf:
		socket.protoBufRouter(conn, message)
	}

}

func (socket *Socket) jsonRouter(conn *Connection, message *Message) {

	var event = gjson.GetBytes(message.Message.([]byte), "event").Str

	var f = socket.GetRouter(event)

	if f == nil {
		return
	}

	f(conn, message, nil)
}

func (socket *Socket) protoBufRouter(conn *Connection, message *Message) {

}
