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

type lobbyReq struct {
	name string
	conn *Connection
}

type managedLobby struct {
	lobby *Lobby
	count uint
}

type LobbyManager struct {
	members  map[*Connection]map[string]bool
	lobbies  map[string]*managedLobby
	join     chan *lobbyReq
	leave    chan *lobbyReq
	leaveAll chan *Connection
	stop     chan bool
}

func NewLobbyManager() *LobbyManager {
	lm := LobbyManager{
		members:  make(map[*Connection]map[string]bool),
		lobbies:  make(map[string]*managedLobby),
		join:     make(chan *lobbyReq),
		leave:    make(chan *lobbyReq),
		leaveAll: make(chan *Connection),
		stop:     make(chan bool),
	}
	go lm.run()
	return &lm
}

func (lm *LobbyManager) leaveLobbyByName(name string, conn *Connection) {
	m, ok := lm.lobbies[name]
	if ok {
		m.lobby.leave <- conn
		m.count--
		if m.count == 0 {
			m.lobby.Stop()
			delete(lm.lobbies, name)
		}
	}
}

func (lm *LobbyManager) run() {
	for {
		select {
		case req := <-lm.join:
			m, ok := lm.lobbies[req.name]
			if !ok {
				m = &managedLobby{
					lobby: NewLobby(),
					count: 1,
				}
			} else {
				m.count++
			}
			m.lobby.join <- req.conn
			lm.members[req.conn][req.name] = true
		case req := <-lm.leave:
			lm.leaveLobbyByName(req.name, req.conn)
		case conn := <-lm.leaveAll:
			if cm, ok := lm.members[conn]; ok {
				for name := range cm {
					lm.leaveLobbyByName(name, conn)
				}
			}
		case <-lm.stop:
			return
		}
	}
}

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
