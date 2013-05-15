package golem

import (
	"code.google.com/p/go.net/websocket"
)

const (
	outChannelSize = 256
)

type Connection struct {
	// The websocket connection.
	socket *websocket.Conn
	// The router it belongs to.
	router *Router
	// Buffered channel of outbound messages.
	out chan string
}

func (conn *Connection) Reader() {
	for {
		var message string
		err := websocket.Message.Receive(conn.socket, &message)
		if err != nil {
			break
		}
		go conn.router.Parse(conn, message) // TODO: test performance and necessity of this goroutine
	}
	conn.CloseSocket()
}

func (conn *Connection) Writer() {
	for message := range conn.out {
		err := websocket.Message.Send(conn.socket, message)
		if err != nil {
			break
		}
	}
	conn.CloseSocket()
}

func (conn *Connection) Send(data interface{}) {
	// TODO
}

func (conn *Connection) CloseSocket() {
	conn.socket.Close()
}
