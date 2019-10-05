package lemo

import "strings"

var globalSocketPath = ""

func (socket *Socket) Group(path string, fn func()) {
	globalSocketPath = path
	fn()
	globalSocketPath = ""
}

func (socket *Socket) SetRouter(path string, f WebSocketServerFunction) {

	if socket.WebSocketRouter == nil {
		socket.WebSocketRouter = make(map[string]WebSocketServerFunction)
	}

	path = globalSocketPath + path

	if socket.IgnoreCase {
		path = strings.ToLower(path)
	}

	socket.WebSocketRouter[path] = f
}

func (socket *Socket) GetRouter(path string) WebSocketServerFunction {

	if socket.WebSocketRouter == nil {
		return nil
	}

	if f, ok := socket.WebSocketRouter[path]; ok {
		return f
	}
	return nil
}

func (socket *Socket) router(conn *Connection, msg *MessagePackage) {

	switch socket.TransportType {
	case Json:
		socket.jsonRouter(conn, msg)
	case ProtoBuf:
		socket.protoBufRouter(conn, msg)
	}

}

func (socket *Socket) jsonRouter(conn *Connection, msg *MessagePackage) {

	var f = socket.GetRouter(msg.Event)

	if f == nil {
		return
	}

	f(conn, msg)
}

func (socket *Socket) protoBufRouter(conn *Connection, msg *MessagePackage) {

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
			return string(bts[s+2 : l-2]), nil
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
