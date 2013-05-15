package golem

import (
	"code.google.com/p/go.net/websocket"
)

type Connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan string
}

func (conn *Connection) Reader() {
	for {
		var message string
		err := websocket.Message.Receive(conn.ws, &message)
		if err != nil {
			break
		}
		hub.broadcast <- message
	}
	conn.CloseSocket()
}

func (conn *Connection) Writer() {
	for message := range conn.send {
		err := websocket.Message.Send(conn.ws, message)
		if err != nil {
			break
		}
	}
	conn.CloseSocket()
}

func (conn *Connection) CloseSocket() {
	conn.ws.Close()
}

func WebSocketHandler(ws *websocket.Conn) {
	conn := &Connection{send: make(chan string, 256), ws: ws}
	hub.register <- conn
	defer func() { hub.unregister <- conn }()
	go conn.Writer()
	conn.Reader()
}
