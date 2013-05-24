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

type Lobby struct {
	members map[*Connection]bool
	stop    chan bool
	join    chan *Connection
	leave   chan *Connection
	send    chan []byte
}

func NewLobby() *Lobby {
	l := Lobby{
		members: make(map[*Connection]bool),
		stop:    make(chan bool),
		join:    make(chan *Connection),
		leave:   make(chan *Connection),
		send:    make(chan []byte),
	}
	go l.run()
	return &l
}

func (l *Lobby) run() {
	for {
		select {
		case conn := <-l.join:
			l.members[conn] = true
		case conn := <-l.leave:
			_, ok := l.members[conn]
			if ok {
				delete(l.members, conn)
			}
		case message := <-l.send:
			for conn := range l.members {
				select {
				case conn.send <- message:
				default:
					delete(l.members, conn)
				}
			}
		case <-l.stop:
			return
		}
	}
}

func (l *Lobby) Remove() {
	l.stop <- true
}

func (l *Lobby) Join(conn *Connection) {
	l.join <- conn
}

func (l *Lobby) Leave(conn *Connection) {
	l.leave <- conn
}

func (l *Lobby) Send(data []byte) {
	l.send <- data
}
