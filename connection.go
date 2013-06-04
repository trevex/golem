/*

   Copyright 2013 Niklas Voss

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

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
)

// Connection holds information about the underlying WebSocket-Connection,
// the associated router and the outgoing data channel.
type Connection struct {
	// The websocket connection.
	socket *websocket.Conn
	// Associated router.
	router *Router
	// Buffered channel of outbound messages.
	send chan *message
}

// Create a new connection using the specified socket and router.
func newConnection(s *websocket.Conn, r *Router) *Connection {
	return &Connection{
		socket: s,
		router: r,
		send:   make(chan *message, sendChannelSize),
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
			if ok {
				if data, err := conn.router.prepareDataForEmit(message.event, message.data); err == nil {
					if err := conn.write(websocket.OpText, data); err != nil {
						return
					}
				}
			} else {
				conn.write(websocket.OpClose, []byte{})
				return
			}
		case <-ticker.C:
			if err := conn.write(websocket.OpPing, []byte{}); err != nil {
				return
			}
		}
	}
}

// Register connection and start writing and reading loops.
func (conn *Connection) run() {
	hub.register <- conn
	go conn.writePump()
	conn.readPump()
}

// Emit event with provided data. The data will be automatically marshalled.
func (conn *Connection) Emit(event string, data interface{}) {
	conn.send <- &message{
		event: event,
		data:  data,
	}
}
