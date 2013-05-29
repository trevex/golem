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

// Lobby holds all channels necessary to handle a single lobby instance.
// The struct should not be directly instanced rather NewLobby() should be used.
type Lobby struct {
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

// Creates and initialised lobby and returns pointer to it.
func NewLobby() *Lobby {
	l := Lobby{
		members: make(map[*Connection]bool),
		stop:    make(chan bool),
		join:    make(chan *Connection),
		leave:   make(chan *Connection),
		send:    make(chan []byte),
	}
	// Run the message loop
	go l.run()
	// Return pointer
	return &l
}

// Starts the message loop of this lobby, should only be run once and in a different routine.
func (l *Lobby) run() {
	for {
		select {
		// Join
		case conn := <-l.join:
			l.members[conn] = true
		// Leave
		case conn := <-l.leave:
			if _, ok := l.members[conn]; ok { // If member exists, delete it
				delete(l.members, conn)
			}
		// Send
		case message := <-l.send:
			for conn := range l.members { // For every connection try to send
				select {
				case conn.send <- message:
				default: // If sending failed, delete member
					delete(l.members, conn)
				}
			}
		// Stop
		case <-l.stop:
			return
		}
	}
}

// Stops the message queue.
func (l *Lobby) Stop() {
	l.stop <- true
}

// The specified connection joins the lobby.
func (l *Lobby) Join(conn *Connection) {
	l.join <- conn
}

// If the specified connection is member of the lobby, the connection will leave it.
func (l *Lobby) Leave(conn *Connection) {
	l.leave <- conn
}

// Send an array of bytes to every member if the channel.
func (l *Lobby) Send(data []byte) {
	l.send <- data
}

// Emits message event to all members of the channel.
func (l *Lobby) Emit(what string, data interface{}) {
	if b, ok := pack(what, data); ok {
		l.send <- b
	}
}
