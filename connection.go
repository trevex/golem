/*

	golem - lightweight Go WebSocket-framework
    Copyright (C) 2013  Niklas Voss

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <http://www.gnu.org/licenses/>.

*/

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
	sendChannelSize = 512
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
	send chan []byte
}

func newConnection(s *websocket.Conn, r *Router) *Connection {
	return &Connection{
		socket: s,
		router: r,
		send:   make(chan []byte, sendChannelSize),
	}
}

// readPump pumps messages from the websocket connection to the hub.
func (conn *Connection) readPump() {
	defer func() {
		hub.unregister <- conn
		conn.socket.Close()
		conn.router.closeCallback(conn)
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
		conn.socket.Close() // Necessary to force reading to stop
	}()
	for {
		select {
		case message, ok := <-conn.send:
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
