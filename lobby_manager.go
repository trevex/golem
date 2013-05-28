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

// Lobby request information holding name of the lobby and the connection, which requested.
type lobbyReq struct {
	// Name of the lobby the request goes to.
	name string
	// Reference to the connection, which requested.
	conn *Connection
}

// Lobby messages contain information about to which lobby it is being send and the data being send.
type lobbyMsg struct {
	// Name of the lobby the message goes to.
	to string
	// Data being send to specified lobby.
	data []byte
}

// Wrapper for normal lobbies to add a member counter.
type managedLobby struct {
	// Reference to lobby.
	lobby *Lobby
	// Member-count to allow removing of empty lobbies.
	count uint
}

// Handles any count of lobbies by keys. Currently only strings are supported as keys (lobby names).
// As soon as generics are supported any key should be able to be used. The methods are used similar to
// single lobby instance but preceded by the key.
type LobbyManager struct {
	// Map of connections mapped to lobbies joined; necessary for leave all/clean up functionality.
	members map[*Connection]map[string]bool
	// Map of all managed lobbies with their names as keys.
	lobbies map[string]*managedLobby
	// Channel of join requests.
	join chan *lobbyReq
	// Channel of leave requests.
	leave chan *lobbyReq
	// Channel of leave all requests, essentially cleaning up every trace of the specified connection.
	leaveAll chan *Connection
	// Channel of messages associated with this lobby manager
	send chan *lobbyMsg
	// Stop signal channel
	stop chan bool
}

// Creates a new LobbyManager-Instance.
func NewLobbyManager() *LobbyManager {
	// Create instance.
	lm := LobbyManager{
		members:  make(map[*Connection]map[string]bool),
		lobbies:  make(map[string]*managedLobby),
		join:     make(chan *lobbyReq),
		leave:    make(chan *lobbyReq),
		leaveAll: make(chan *Connection),
		send:     make(chan *lobbyMsg),
		stop:     make(chan bool),
	}
	// Start message loop in new routine.
	go lm.run()
	// Return reference to this lobby manager.
	return &lm
}

// Helper function to leave a lobby by name. If specified lobby has
// no members after leaving, it will be cleaned up.
func (lm *LobbyManager) leaveLobbyByName(name string, conn *Connection) {
	if m, ok := lm.lobbies[name]; ok { // Continue if getting the lobby was ok.
		if _, ok := lm.members[conn]; ok { // Continue if connection has map of joined lobbies.
			if _, ok := lm.members[conn][name]; ok { // Continue if connection actually joined specified lobby.
				m.lobby.leave <- conn
				m.count--
				delete(lm.members[conn], name)
				if m.count == 0 { // Get rid of lobby if it is empty
					m.lobby.Stop()
					delete(lm.lobbies, name)
				}
			}
		}
	}
}

// Run should always be executed in a new goroutine, because it contains the
// message loop.
func (lm *LobbyManager) run() {
	for {
		select {
		// Join
		case req := <-lm.join:
			m, ok := lm.lobbies[req.name]
			if !ok { // If lobby was not found for join request, create it!
				m = &managedLobby{
					lobby: NewLobby(),
					count: 1, // start with count 1 for first user
				}
				lm.lobbies[req.name] = m
			} else { // If lobby exists increase count and join.
				m.count++
			}
			m.lobby.join <- req.conn
			if _, ok := lm.members[req.conn]; !ok { // If lobby association map for connection does not exist, create it!
				lm.members[req.conn] = make(map[string]bool)
			}
			lm.members[req.conn][req.name] = true // Flag this lobby on members lobby map.
		// Leave
		case req := <-lm.leave:
			lm.leaveLobbyByName(req.name, req.conn)
		// Leave all
		case conn := <-lm.leaveAll:
			if cm, ok := lm.members[conn]; ok {
				for name := range cm { // Iterate over all lobbies this connection joined and leave them.
					lm.leaveLobbyByName(name, conn)
				}
				delete(lm.members, conn) // Remove map of joined lobbies
			}
		// Send
		case msg := <-lm.send:
			if m, ok := lm.lobbies[msg.to]; ok { // If lobby exists, get it and send data to it.
				m.lobby.send <- msg.data
			}
		// Stop
		case <-lm.stop:
			for k, m := range lm.lobbies { // Stop all lobbies!
				m.lobby.Stop()
				delete(lm.lobbies, k)
			}
			return
		}
	}
}

//
func (lm *LobbyManager) Join(name string, conn *Connection) {
	lm.join <- &lobbyReq{
		name: name,
		conn: conn,
	}
}

func (lm *LobbyManager) Leave(name string, conn *Connection) {
	lm.leave <- &lobbyReq{
		name: name,
		conn: conn,
	}
}

func (lm *LobbyManager) LeaveAll(conn *Connection) {
	lm.leaveAll <- conn
}

func (lm *LobbyManager) Send(to string, data []byte) {
	lm.send <- &lobbyMsg{
		to:   to,
		data: data,
	}
}

func (lm *LobbyManager) Emit(to string, what string, data interface{}) {
	if b, ok := pack(what, data); ok {
		lm.send <- &lobbyMsg{
			to:   to,
			data: b,
		}
	}
}

func (lm *LobbyManager) Stop() {
	lm.stop <- true
}
