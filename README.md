    package main
    
    import (
        "log"
    
        "github.com/Lemo-yxk/ws"
    )
    
    func init() {
        log.SetFlags(log.Llongfile | log.Ldate | log.Ltime)
    }
    
    func main() {
    
        var server = &lemo.Server{Host: "127.0.0.1", Port: 12345, Path: "/Game-Robot"}
    
        var socketHandler = &lemo.Socket{}
    
        socketHandler.SetRouter("hello1", func(conn *lemo.Connection, ftd *lemo.Fte, msg []byte) {
            log.Println(ftd.Fd)
        })
    
        socketHandler.OnClose = func(conn *lemo.Connection) {
            log.Println(conn.Fd, "is close")
        }
    
        socketHandler.OnError = func(err error) {
            log.Println(err)
        }
    
        socketHandler.OnOpen = func(conn *lemo.Connection) {
            log.Println(conn.Fd, "is open")
        }
    
        var httpHandler = &lemo.Http{}
    
        httpHandler.Group("/hello", func() {
            httpHandler.Get("/:name", func(t *lemo.Stream) {
                log.Println(t.Params)
                _ = t.End(t.Params.ByName("name"))
            })
        })
    
        server.Start(socketHandler, httpHandler)
    
    }
