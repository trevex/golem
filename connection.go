package golem

import (
	"github.com/garyburd/go-websocket/websocket"
	"io/ioutil"
	"time"
)

const (
	// Time allowed to write a message to the client.
	writeWait = 10 * time.Second
	// Time allowed to read the next message from the client.
	readWait = 60 * time.Second
	// Send pings to client with this period. Must be less than readWait.
	pingPeriod = (readWait * 9) / 10
	// Maximum message size allowed from client.
	maxMessageSize = 512
	// Outgoing default channel size.
	outChannelSize = 512
	// Default lobby capacity
	lobbyDefaultCapacity = 3
)

// connection is an middleman between the websocket connection and the hub.
type Connection struct {
	// The websocket connection.
	socket *websocket.Conn
	// Associated router.
	router *Router
	// Buffered channel of outbound messages.
	out chan []byte
}

func newConnection(s *websocket.Conn, r *Router) *Connection {
	return &Connection{
		socket: s,
		router: r,
		out:    make(chan []byte, outChannelSize),
	}
}

// readPump pumps messages from the websocket connection to the hub.
func (conn *Connection) readPump() {
	defer func() {
		hub.connMgr.unregister <- conn
		conn.socket.Close()
	}()
	conn.socket.SetReadLimit(maxMessageSize)
	conn.socket.SetReadDeadline(time.Now().Add(readWait))
	for {
		op, r, err := conn.socket.NextReader()
		if err != nil {
			break
		}
		switch op {
		case websocket.OpPong:
			conn.socket.SetReadDeadline(time.Now().Add(readWait))
		case websocket.OpText:
			message, err := ioutil.ReadAll(r)
			if err != nil {
				break
			}
			conn.router.parse(conn, message)
		}
	}
}

// write writes a message with the given opCode and payload.
func (conn *Connection) write(opCode int, payload []byte) error {
	conn.socket.SetWriteDeadline(time.Now().Add(writeWait))
	return conn.socket.WriteMessage(opCode, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (conn *Connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.socket.Close()
	}()
	for {
		select {
		case message, ok := <-conn.out:
			if !ok {
				conn.write(websocket.OpClose, []byte{})
				return
			}
			if err := conn.write(websocket.OpText, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := conn.write(websocket.OpPing, []byte{}); err != nil {
				return
			}
		}
	}
}

//
func (conn *Connection) Send(msg interface{}) {

}
