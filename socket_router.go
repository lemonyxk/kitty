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

func (socket *Socket) router(conn *Connection, ftd *Fte, msg []byte) {

	switch socket.TsProto {
	case Json:
		socket.jsonRouter(conn, ftd, msg)
	case ProtoBuf:
		socket.protoBufRouter(conn, ftd, msg)
	}

}

func (socket *Socket) jsonRouter(conn *Connection, fte *Fte, msg []byte) {

	if len(msg) < 22 {
		return
	}

	var event = gjson.GetBytes(msg, "event").Str

	var data = []byte(gjson.GetBytes(msg, "data").Str)

	var f = socket.GetRouter(event)

	if f == nil {
		return
	}

	fte.Event = event

	f(conn, fte, data)
}

func (socket *Socket) protoBufRouter(conn *Connection, fte *Fte, msg []byte) {

}

func ParseMessage(bts []byte) (string, []byte) {

	var s, e int

	var l = len(bts)

	// 正序
	if bts[8] == 58 {

		s = 8

		for i, b := range bts {
			if b == 44 {
				e = i
				break
			}
		}

		if e == 0 {
			return string(bts[s+2:]), nil
		}

		return string(bts[s+2 : e-1]), bts[e+8 : l-1]

	} else {

		for i := l - 1; i >= 0; i-- {

			if bts[i] == 58 {
				s = i
			}

			if bts[i] == 44 {
				e = i
				break
			}
		}

		if s == 0 {
			return "", nil
		}

		if e == 0 {
			return string(bts[s+2 : l-2]), nil
		}

		return string(bts[s+2 : l-2]), bts[8:e]
	}
}
