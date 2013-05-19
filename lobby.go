package golem

type lobby struct {
	// Connection registered to this lobby.
	connections map[*Connection]bool

	// Message that should be send to subscribed Connections.
	broadcast chan []byte

	// Subscribe to lobby.
	subscribe chan *Connection

	// Unsubscribe Connection from lobby.
	unsubscribe chan *Connection

	manager *lobbyManager
}

func newLobby(mngr *lobbyManager) *lobby {
	return &lobby{
		broadcast:   make(chan []byte),
		subscribe:   make(chan *Connection),
		unsubscribe: make(chan *Connection),
		connections: make(map[*Connection]bool),
		manager:     mngr,
	}
}

func (l *lobby) run() {
	for {
		select {
		case conn := <-l.subscribe:
			l.connections[conn] = true
		case conn := <-l.unsubscribe:
			if !l.removeAndCheck(conn) {
				return
			}
		case message := <-l.broadcast:
			for conn := range l.connections {
				select {
				case conn.out <- message:
				default:
					if !l.removeAndCheck(conn) {
						return
					}
				}
			}
		}
	}
}

func (l *lobby) removeAndCheck(conn *Connection) bool {
	delete(l.connections, conn)
	if len(l.connections) == 0 {
		l.manager.remove <- l
		close(l.broadcast)
		close(l.subscribe)
		close(l.unsubscribe)
		return false
	}
	return true
}
