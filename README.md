    package main
    
    import (
        "github.com/Lemo-yxk/ws"
        "log"
    )
    
    func init() {
        log.SetFlags(log.Llongfile | log.Ldate | log.Ltime)
    }
    
    func main() {
    
        var server = &ws.Server{Host: "127.0.0.1", Port: 5858, Path: "/Game-Robot"}
    
        var socketHandler = &ws.Socket{}
    
        socketHandler.SetRouter("hello1", func(conn *ws.Connection, ftd *ws.Fte, msg []byte) {
            log.Println(ftd.Fd)
        })
    
        socketHandler.OnClose = func(conn *ws.Connection) {
            log.Println(conn.Fd, "is close")
        }
    
        socketHandler.OnError = func(err error) {
            log.Println(err)
        }
    
        socketHandler.OnOpen = func(conn *ws.Connection) {
            log.Println(conn.Fd, "is open")
        }
    
        var httpHandler = &ws.HttpHandle{}
    
        httpHandler.Group("/hello", []ws.Before{
            func(t *ws.Stream) (i interface{}, e error) {
                log.Println("before")
                return nil, nil
            },
        }, func() {
            httpHandler.Get("/xixi", func(t *ws.Stream) {
                log.Println("now")
                _ = t.Json("hello2")
            })
        }, []ws.After{
            func(t *ws.Stream) error {
                log.Println("after")
                return nil
            },
        })
    
        server.Start(ws.WebSocket(socketHandler), httpHandler)
    
    }
