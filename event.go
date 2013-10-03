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
	return BroadcastServer{sockets: make(chan *client),
		Broadcast: make(chan interface{}),
		clients:   make(map[string]*list.List),
		dead:      make(chan *client)}
}

func TagEventServer(f TagGen) BroadcastServer {
	return BroadcastServer{sockets: make(chan *client),
		Broadcast: make(chan interface{}),
		Tags:      make(chan TaggedMessage),
		clients:   make(map[string]*list.List),
		gentag:    f,
		dead:      make(chan *client)}
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
			b.broadcastTag("", msg)

		case msg := <-b.Tags:
			b.broadcastTag(msg.Tag, msg.Msg)

		case client := <-b.dead:
			l := b.clients[client.tag]
			l.Remove(client.node)
		}
	}
}

func (b *BroadcastServer) broadcastTag(tag string, msg interface{}) {
	l, ok := b.clients[tag]
	if !ok {
		return
	}

	data, err := json.Marshal(msg)
	check(err)

	for cur := l.Front(); cur != nil; cur = cur.Next() {
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

		buf.Write([]byte("HTTP/1.1 200 OK\r\n"))
		buf.Write([]byte("Content-Type: text/event-stream\r\n"))
		buf.Write([]byte("X-Accel-Buffering: no\r\n\r\n"))
		check(buf.Flush())

		/* Don't clog the system too much if one particular write is slow by holding
		   a buffer of a few messages */
		msgs := make(chan []byte, 10)
		c := client{msgs: msgs, tag: tag}
		b.sockets <- &c
		dead := make(chan int)

		/* spawn off something to read and close the connection when they get
		   disconnected to prevent lots of things lying around */
		go func() {
			buf := []byte{0, 0, 0, 0}
			for {
				_, err := conn.Read(buf)
				if err != nil {
					dead <- 1
					break
				}
			}
		}()

		defer func() {
			b.dead <- &c
			conn.Close()
		}()

		for {
			select {
			case <-dead:
				return

			case msg := <-msgs:
				_, err := buf.Write([]byte("data: "))
				if err == nil {
					_, err = buf.Write(msg)
				}
				if err == nil {
					_, err = buf.Write([]byte("\n\n"))
				}
				if err == nil {
					err = buf.Flush()
				}
				if err != nil {
					return
				}
			}
		}
	})
}
