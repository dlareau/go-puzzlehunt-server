package main

import "container/list"
import "net/http"
import ws "code.google.com/p/go.net/websocket"

type TagGen func(*http.Request) string

type BroadcastServer struct {
  sockets   chan *client
  clients   map[string]*list.List
  Broadcast chan interface{}
  Tags      chan TaggedMessage
  gentag    TagGen
}

type TaggedMessage struct {
  Tag string
  Msg interface{}
}

type client struct {
  ws   *ws.Conn
  done chan int
  tag  string
}

func WsServer() BroadcastServer {
  return BroadcastServer{ sockets: make(chan *client),
                          Broadcast: make(chan interface{}),
                          clients: make(map[string]*list.List) }
}

func TagWsServer(f TagGen) BroadcastServer {
  return BroadcastServer{ sockets: make(chan *client),
                          Broadcast: make(chan interface{}),
                          Tags: make(chan TaggedMessage),
                          clients: make(map[string]*list.List),
                          gentag: f }
}

func (b *BroadcastServer) Serve() {
  for {
    select {
      case ws := <-b.sockets:
        l, ok := b.clients[ws.tag]
        if !ok {
          l = list.New()
          b.clients[ws.tag] = l
        }
        l.PushBack(ws)

      case msg := <-b.Broadcast:
        b.broadcastTag("", msg);

      case msg := <-b.Tags:
        b.broadcastTag(msg.Tag, msg.Msg);
    }
  }
}

func (b *BroadcastServer) broadcastTag(tag string, msg interface{}) {
  l, ok := b.clients[tag]
  if !ok { return }
  var nxt *list.Element
  for cur := l.Front(); cur != nil; cur = nxt {
    nxt = cur.Next()
    client := cur.Value.(*client)
    if ws.JSON.Send(client.ws, msg) != nil {
      l.Remove(cur)
      client.done <- 1
    }
  }
}

func (b *BroadcastServer) Endpoint() http.Handler {
  return ws.Handler(func(ws *ws.Conn) {
    tag := ""
    if b.gentag != nil {
      tag = b.gentag(ws.Request())
    }
    c := client{ ws: ws, done: make(chan int), tag: tag }
    b.sockets <- &c
    <-c.done
  })
}
