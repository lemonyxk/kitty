package ws

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

func (socket *Socket) router(conn *Connection, message *BM) {

	switch socket.TsProto {
	case Json:
		socket.jsonRouter(conn, message)
	case ProtoBuf:
		socket.protoBufRouter(conn, message)
	}

}

func (socket *Socket) jsonRouter(conn *Connection, message *BM) {

	var event, data = parseMessage(message.Msg)

	var f = socket.GetRouter(event)

	if f == nil {
		return
	}

	message.Event = event
	message.Msg = data

	f(conn, message, nil)
}

func (socket *Socket) protoBufRouter(conn *Connection, message *BM) {

}

func parseMessage(bts []byte) (string, []byte) {

	var s, e int

	for i, b := range bts {
		if b == 58 {
			s = i
		}
		if b == 44 {
			e = i
			break
		}
	}

	if s == 0 {
		return "", nil
	}

	if e == 0 {
		return string(bts[s+2:]), nil
	}

	return string(bts[s+2 : e-1]), bts[e+8 : len(bts)-1]
}
