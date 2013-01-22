package main

import "container/list"
import "net/http"
import ws "code.google.com/p/go.net/websocket"

type BroadcastServer struct {
  sockets  chan *client
  Messages chan interface{}
}

type client struct {
  ws   *ws.Conn
  done chan int
}

func WsServer() BroadcastServer {
  return BroadcastServer{ sockets: make(chan *client),
                          Messages: make(chan interface{}) }
}

func (b *BroadcastServer) Serve() {
  sockets := list.New()
  for {
    select {
      case ws := <-b.sockets:
        sockets.PushBack(ws)

      case msg := <-b.Messages:
        var nxt *list.Element
        for cur := sockets.Front(); cur != nil; cur = nxt {
          nxt = cur.Next()
          client := cur.Value.(*client)
          if ws.JSON.Send(client.ws, msg) != nil {
            sockets.Remove(cur)
            client.done <- 1
          }
        }
    }
  }
}

func (b *BroadcastServer) Endpoint() http.Handler {
  return ws.Handler(func(ws *ws.Conn) {
    c := client{ ws: ws, done: make(chan int) }
    b.sockets <- &c
    <-c.done
  })
}
