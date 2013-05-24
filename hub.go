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

func (hub *Hub) remove(conn *Connection) {
	delete(hub.connections, conn)
	close(conn.send)
}

func (hub *Hub) run() {
	if hub.isRunning != true {
		hub.isRunning = true
		go func() {
			for {
				select {
				case conn := <-hub.register:
					hub.connections[conn] = true
				case conn := <-hub.unregister:
					hub.remove(conn)
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

var hub = Hub{
	broadcast:   make(chan []byte),
	register:    make(chan *Connection),
	unregister:  make(chan *Connection),
	connections: make(map[*Connection]bool),
	isRunning:   false,
}

func GetHub() *Hub {
	return &hub
}
