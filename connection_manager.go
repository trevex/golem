package golem

type connectionManager struct {
	// Registered connections.
	connections map[*Connection]bool

	// Inbound messages from the connections.
	broadcast chan []byte

	// Register requests from the connections.
	register chan *Connection

	// Unregister requests from connections.
	unregister chan *Connection
}

func newConnectionManager() *connectionManager {
	return &connectionManager{
		broadcast:   make(chan []byte),
		register:    make(chan *Connection),
		unregister:  make(chan *Connection),
		connections: make(map[*Connection]bool),
	}
}

func (cm *connectionManager) remove(conn *Connection) {
	delete(cm.connections, conn)
	close(conn.out)
}

func (cm *connectionManager) run() {
	for {
		select {
		case conn := <-cm.register:
			cm.connections[conn] = true
		case conn := <-cm.unregister:
			cm.remove(conn)
		case message := <-cm.broadcast:
			for conn := range cm.connections {
				select {
				case conn.out <- message:
				default:
					cm.remove(conn)
				}
			}
		}
	}
}
