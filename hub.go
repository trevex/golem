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

// The hub manages all active connection, but should only be used directly
// if broadcasting of data or an event is desired. The hub should not be instanced
// directly.
type Hub struct {
	// Registered connections.
	connections map[*Connection]bool

	// Inbound messages from the connections.
	broadcast chan []byte

	// Register requests from the connections.
	register chan *Connection

	// Unregister requests from connections.
	unregister chan *Connection

	// Flag to determine if running or not
	isRunning bool
}

// Remove the specified connection from the hub and drop the socket.
func (hub *Hub) remove(conn *Connection) {
	delete(hub.connections, conn)
	close(conn.send)
}

// If the hub is not running, start it in a different goroutine.
func (hub *Hub) run() {
	if hub.isRunning != true { // Should be safe, because only called from NewRouter and therefore a single thread.
		hub.isRunning = true
		go func() {
			for {
				select {
				// Register new connection
				case conn := <-hub.register:
					hub.connections[conn] = true
				// Unregister dropped connection
				case conn := <-hub.unregister:
					hub.remove(conn)
				// Broadcast
				case message := <-hub.broadcast:
					for conn := range hub.connections {
						select {
						case conn.send <- message:
						default:
							hub.remove(conn)
						}
					}
				}
			}
		}()
	}
}

// Create the hub instance.
var hub = Hub{
	broadcast:   make(chan []byte),
	register:    make(chan *Connection),
	unregister:  make(chan *Connection),
	connections: make(map[*Connection]bool),
	isRunning:   false,
}

// Retrieve pointer to the hub.
func GetHub() *Hub {
	return &hub
}

// Broadcast an array of bytes to all active connections.
func (hub *Hub) Broadcast(data []byte) {
	hub.broadcast <- data
}

// Broadcast event to all active connections.
func (hub *Hub) BroadcastEmit(what string, data interface{}) {
	if b, ok := pack(what, data); ok {
		hub.broadcast <- b
	}
}
