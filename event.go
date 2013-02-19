package main

import "encoding/json"
import "container/list"
import "net/http"

type TagGen func(*http.Request) string

type BroadcastServer struct {
  sockets   chan *client
  dead      chan *client
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
  node *list.Element
  msgs chan []byte
  tag  string
}

func EventServer() BroadcastServer {
  return BroadcastServer{ sockets: make(chan *client),
                          Broadcast: make(chan interface{}),
                          clients: make(map[string]*list.List),
                          dead: make(chan *client)}
}

func TagEventServer(f TagGen) BroadcastServer {
  return BroadcastServer{ sockets: make(chan *client),
                          Broadcast: make(chan interface{}),
                          Tags: make(chan TaggedMessage),
                          clients: make(map[string]*list.List),
                          gentag: f,
                          dead: make(chan *client) }
}

func (b *BroadcastServer) Serve() {
  for {
    select {
      case client := <-b.sockets:
        l, ok := b.clients[client.tag]
        if !ok {
          l = list.New()
          b.clients[client.tag] = l
        }
        l.PushBack(client)
        client.node = l.Back()

      case msg := <-b.Broadcast:
        b.broadcastTag("", msg);

      case msg := <-b.Tags:
        b.broadcastTag(msg.Tag, msg.Msg);

      case client := <-b.dead:
        l := b.clients[client.tag]
        l.Remove(client.node)
    }
  }
}

func (b *BroadcastServer) broadcastTag(tag string, msg interface{}) {
  data, err := json.Marshal(msg)
  check(err)

  l, ok := b.clients[tag]
  if !ok { return }
  var nxt *list.Element
  for cur := l.Front(); cur != nil; cur = nxt {
    nxt = cur.Next()
    client := cur.Value.(*client)
    client.msgs <- data
  }
}

func (b *BroadcastServer) Endpoint() http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    tag := ""
    if b.gentag != nil {
      tag = b.gentag(r)
    }
    hj := w.(http.Hijacker)
    conn, buf, err := hj.Hijack()
    check(err)
    defer conn.Close()

    buf.Write([]byte("HTTP/1.1 200 OK\r\n"))
    buf.Write([]byte("Content-Type: text/event-stream\r\n"))
    buf.Write([]byte("X-Accel-Buffering: no\r\n\r\n"))
    check(buf.Flush())

    msgs := make(chan []byte)
    c := client{ msgs: msgs, tag: tag }
    b.sockets <- &c
    for msg := range msgs {
      _, err := buf.Write([]byte("data: "))
      if err == nil { _, err = buf.Write(msg) }
      if err == nil { _, err = buf.Write([]byte("\n\n")) }
      if err == nil { err = buf.Flush() }
      if err != nil {
        b.dead <- &c
      }
    }
  })
}
