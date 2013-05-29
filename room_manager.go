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

// Room request information holding name of the room and the connection, which requested.
type roomReq struct {
	// Name of the lobby the request goes to.
	name string
	// Reference to the connection, which requested.
	conn *Connection
}

// Room messages contain information about to which room it is being send and the data being send.
type roomMsg struct {
	// Name of the room the message goes to.
	to string
	// Data being send to specified room.
	data []byte
}

// Wrapper for normal lobbies to add a member counter.
type managedRoom struct {
	// Reference to room.
	room *Room
	// Member-count to allow removing of empty lobbies.
	count uint
}

// Handles any count of lobbies by keys. Currently only strings are supported as keys (room names).
// As soon as generics are supported any key should be able to be used. The methods are used similar to
// single room instance but preceded by the key.
type RoomManager struct {
	// Map of connections mapped to lobbies joined; necessary for leave all/clean up functionality.
	members map[*Connection]map[string]bool
	// Map of all managed lobbies with their names as keys.
	rooms map[string]*managedRoom
	// Channel of join requests.
	join chan *roomReq
	// Channel of leave requests.
	leave chan *roomReq
	// Channel of leave all requests, essentially cleaning up every trace of the specified connection.
	leaveAll chan *Connection
	// Channel of messages associated with this room manager
	send chan *roomMsg
	// Stop signal channel
	stop chan bool
}

// Creates a new RoomManager-Instance.
func NewRoomManager() *RoomManager {
	// Create instance.
	rm := RoomManager{
		members:  make(map[*Connection]map[string]bool),
		rooms:    make(map[string]*managedRoom),
		join:     make(chan *roomReq),
		leave:    make(chan *roomReq),
		leaveAll: make(chan *Connection),
		send:     make(chan *roomMsg),
		stop:     make(chan bool),
	}
	// Start message loop in new routine.
	go rm.run()
	// Return reference to this room manager.
	return &rm
}

// Helper function to leave a room by name. If specified room has
// no members after leaving, it will be cleaned up.
func (rm *RoomManager) leaveRoomByName(name string, conn *Connection) {
	if m, ok := rm.rooms[name]; ok { // Continue if getting the room was ok.
		if _, ok := rm.members[conn]; ok { // Continue if connection has map of joined lobbies.
			if _, ok := rm.members[conn][name]; ok { // Continue if connection actually joined specified room.
				m.room.leave <- conn
				m.count--
				delete(rm.members[conn], name)
				if m.count == 0 { // Get rid of room if it is empty
					m.room.Stop()
					delete(rm.rooms, name)
				}
			}
		}
	}
}

// Run should always be executed in a new goroutine, because it contains the
// message loop.
func (rm *RoomManager) run() {
	for {
		select {
		// Join
		case req := <-rm.join:
			m, ok := rm.rooms[req.name]
			if !ok { // If room was not found for join request, create it!
				m = &managedRoom{
					room:  NewRoom(),
					count: 1, // start with count 1 for first user
				}
				rm.rooms[req.name] = m
			} else { // If room exists increase count and join.
				m.count++
			}
			m.room.join <- req.conn
			if _, ok := rm.members[req.conn]; !ok { // If room association map for connection does not exist, create it!
				rm.members[req.conn] = make(map[string]bool)
			}
			rm.members[req.conn][req.name] = true // Flag this room on members room map.
		// Leave
		case req := <-rm.leave:
			rm.leaveRoomByName(req.name, req.conn)
		// Leave all
		case conn := <-rm.leaveAll:
			if cm, ok := rm.members[conn]; ok {
				for name := range cm { // Iterate over all lobbies this connection joined and leave them.
					rm.leaveRoomByName(name, conn)
				}
				delete(rm.members, conn) // Remove map of joined lobbies
			}
		// Send
		case msg := <-rm.send:
			if m, ok := rm.rooms[msg.to]; ok { // If room exists, get it and send data to it.
				m.room.send <- msg.data
			}
		// Stop
		case <-rm.stop:
			for k, m := range rm.rooms { // Stop all lobbies!
				m.room.Stop()
				delete(rm.rooms, k)
			}
			return
		}
	}
}

// The connection joins the room with the specified name.
func (rm *RoomManager) Join(name string, conn *Connection) {
	rm.join <- &roomReq{
		name: name,
		conn: conn,
	}
}

// The connection leaves the room with the specified name.
func (rm *RoomManager) Leave(name string, conn *Connection) {
	rm.leave <- &roomReq{
		name: name,
		conn: conn,
	}
}

// The connection leaves all lobbies of this manager. This is important for clean up purposes to
// keep the member count accurate. This should therefore always be called when a connection is closed.
func (rm *RoomManager) LeaveAll(conn *Connection) {
	rm.leaveAll <- conn
}

// Send an array of bytes to all members of the room with the specified name.
func (rm *RoomManager) Send(to string, data []byte) {
	rm.send <- &roomMsg{
		to:   to,
		data: data,
	}
}

// Emit a message, that can be fetched using the golem client library. The provided
// data interface will be automatically marshalled into JSON.
func (rm *RoomManager) Emit(to string, what string, data interface{}) {
	if b, ok := pack(what, data); ok {
		rm.send <- &roomMsg{
			to:   to,
			data: b,
		}
	}
}

// Stop the message loop.
func (rm *RoomManager) Stop() {
	rm.stop <- true
}
