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

// A room is a group of connections, to allow broadcasting to groups.
//
type Room struct {
	// Map of member connections
	members map[*Connection]bool
	// Stop channel
	stop chan bool
	// Join request
	join chan *Connection
	// Leave request
	leave chan *Connection
	// Broadcast to channel data
	send chan []byte
}

// Creates and initialised a room and returns pointer to it.
func NewRoom() *Room {
	r := Room{
		members: make(map[*Connection]bool),
		stop:    make(chan bool),
		join:    make(chan *Connection),
		leave:   make(chan *Connection),
		send:    make(chan []byte),
	}
	// Run the message loop
	go r.run()
	// Return pointer
	return &r
}

// Starts the message loop of this room, should only be run once and in a different routine.
func (r *Room) run() {
	for {
		select {
		// Join
		case conn := <-r.join:
			r.members[conn] = true
		// Leave
		case conn := <-r.leave:
			if _, ok := r.members[conn]; ok { // If member exists, delete it
				delete(r.members, conn)
			}
		// Send
		case message := <-r.send:
			for conn := range r.members { // For every connection try to send
				select {
				case conn.send <- message:
				default: // If sending failed, delete member
					delete(r.members, conn)
				}
			}
		// Stop
		case <-r.stop:
			return
		}
	}
}

// Stops the message queue.
func (r *Room) Stop() {
	r.stop <- true
}

// The specified connection joins the room.
func (r *Room) Join(conn *Connection) {
	r.join <- conn
}

// If the specified connection is member of the room, the connection will leave it.
func (r *Room) Leave(conn *Connection) {
	r.leave <- conn
}

// Send an array of bytes to every member of the channel.
func (r *Room) Send(data []byte) {
	r.send <- data
}

// Emits message event to all members of the channel.
func (r *Room) Emit(what string, data interface{}) {
	if b, ok := pack(what, data); ok {
		r.send <- b
	}
}
