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

const (
	roomManagerCreateEvent        = "create"
	roomManagerRemoveEvent        = "remove"
	CloseConnectionOnLastRoomLeft = 1
)

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
	msg *message
}

// Wrapper for normal lobbies to add a member counter.
type managedRoom struct {
	// Reference to room.
	room *Room
	// Member-count to allow removing of empty lobbies.
	count uint
}

// Structure containing all necessary informations and options of
// connection for the room manager instance
type connectionInfo struct {
	rooms   map[string]bool
	options uint32
}

type connectionInfoReq struct {
	conn      *Connection
	options   uint32
	overwrite bool
}

// Constructor for connection info struct
func newConnectionInfo() *connectionInfo {
	return &connectionInfo{
		rooms:   make(map[string]bool),
		options: 0,
	}
}

// Handles any count of lobbies by keys. Currently only strings are supported as keys (room names).
// As soon as generics are supported any key should be able to be used. The methods are used similar to
// single rooms but preceded by the key.
type RoomManager struct {
	// Map of connections mapped to lobbies joined; necessary for leave all/clean up functionality.
	members map[*Connection]*connectionInfo
	// Map of all managed lobbies with their names as keys.
	rooms map[string]*managedRoom
	// Channel of join requests.
	join chan *roomReq
	// Channel of leave requests.
	leave chan *roomReq
	// Channel of leave all requests, essentially cleaning up every trace of the specified connection.
	leaveAll chan *Connection
	// Channel of room destruction requests
	destroy chan string
	// Channel of connection option requests
	options chan *connectionInfoReq
	// Channel of messages associated with this room manager
	send chan *roomMsg
	// Stop signal channel
	stop chan bool
	// Room creation and removal callbacks
	callbackRoomCreation func(string)
	callbackRoomRemoval  func(string)
}

// NewRoomManager initialises a new instance and returns the a pointer to it.
func NewRoomManager() *RoomManager {
	// Create instance.
	rm := RoomManager{
		members:              make(map[*Connection]*connectionInfo),
		rooms:                make(map[string]*managedRoom),
		join:                 make(chan *roomReq),
		leave:                make(chan *roomReq),
		leaveAll:             make(chan *Connection),
		destroy:              make(chan string),
		options:              make(chan *connectionInfoReq),
		send:                 make(chan *roomMsg, roomSendChannelSize),
		stop:                 make(chan bool),
		callbackRoomCreation: func(string) {},
		callbackRoomRemoval:  func(string) {},
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
		if c, ok := rm.members[conn]; ok { // Continue if connection has map of joined lobbies.
			if _, ok := c.rooms[name]; ok { // Continue if connection actually joined specified room.
				m.room.leave <- conn
				m.count--
				delete(c.rooms, name)
				if len(c.rooms) == 0 && (c.options&CloseConnectionOnLastRoomLeft) == CloseConnectionOnLastRoomLeft {
					delete(rm.members, conn)
					conn.Close()
				}
				if m.count == 0 { // Get rid of room if it is empty
					m.room.Stop()
					delete(rm.rooms, name)
					go rm.callbackRoomRemoval(name)
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
				go rm.callbackRoomCreation(req.name)
			} else { // If room exists increase count and join.
				m.count++
			}
			m.room.join <- req.conn
			c, ok := rm.members[req.conn]
			if !ok { // If room association map for connection does not exist, create it!
				c = newConnectionInfo()
				rm.members[req.conn] = c
			}
			c.rooms[req.name] = true // Flag this room on members room map.
		// Leave
		case req := <-rm.leave:
			rm.leaveRoomByName(req.name, req.conn)
		// Leave all
		case conn := <-rm.leaveAll:
			if c, ok := rm.members[conn]; ok {
				for name := range c.rooms { // Iterate over all lobbies this connection joined and leave them.
					rm.leaveRoomByName(name, conn)
				}
				delete(rm.members, conn) // Remove map of joined lobbies
			}
		case name := <-rm.destroy:
			if m, ok := rm.rooms[name]; ok {
				// This should result inthe room being stopped/destroyed when the last
				// connection is dropped
				for conn := range m.room.members {
					rm.leaveRoomByName(name, conn)
				}
			}
		case req := <-rm.options:
			c, ok := rm.members[req.conn]
			if !ok { // If room association map for connection does not exist, create it!
				c = newConnectionInfo()
				rm.members[req.conn] = c
			}
			if req.overwrite {
				c.options = req.options
			} else {
				c.options = req.options | c.options
			}
		// Send
		case rMsg := <-rm.send:
			if m, ok := rm.rooms[rMsg.to]; ok { // If room exists, get it and send data to it.
				m.room.send <- rMsg.msg
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

func (rm *RoomManager) SetConnectionOptions(conn *Connection, options uint32, overwrite bool) {
	rm.options <- &connectionInfoReq{
		conn:      conn,
		options:   options,
		overwrite: overwrite,
	}
}

// Join adds the connection to the specified room.
func (rm *RoomManager) Join(name string, conn *Connection) {
	rm.join <- &roomReq{
		name: name,
		conn: conn,
	}
}

// Leave removes the connection from the specified room.
func (rm *RoomManager) Leave(name string, conn *Connection) {
	rm.leave <- &roomReq{
		name: name,
		conn: conn,
	}
}

// LeaveAll removes the connection from all joined rooms of this manager.
// This is an important step and should be called OnClose for all connections, that could have joined
// a room of the manager, to keep the member reference count of the manager accurate.
func (rm *RoomManager) LeaveAll(conn *Connection) {
	rm.leaveAll <- conn
}

// Emit a message, that can be fetched using the golem client library. The provided
// data interface will be automatically marshalled according to the active protocol.
func (rm *RoomManager) Emit(to string, event string, data interface{}) {
	rm.send <- &roomMsg{
		to: to,
		msg: &message{
			event: event,
			data:  data,
		},
	}
}

// Stop the message loop and shutsdown the manager. It is safe to delete the instance afterwards.
func (rm *RoomManager) Stop() {
	rm.stop <- true
}

// Remove connections from a particular room and delete the room
func (rm *RoomManager) Destroy(name string) {
	rm.destroy <- name
}

// The room manager can emit several events. At the moment there are two events:
// "create" - triggered if a room was created and
// "remove" - triggered when a room was removed because of insufficient users
// For both the callback needs to be of the type func(string) where the argument
func (rm *RoomManager) On(eventName string, callback interface{}) {
	switch eventName {
	case roomManagerCreateEvent:
		rm.callbackRoomCreation = callback.(func(string))
	case roomManagerRemoveEvent:
		rm.callbackRoomRemoval = callback.(func(string))
	}
}
