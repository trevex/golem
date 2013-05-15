package golem

type Hub struct {
	// Registered connections.
	connections map[*Connection]bool

	// Inbound messages from the connections.
	broadcast chan string

	// Register requests from the connections.
	register chan *Connection

	// Unregister requests from connections.
	unregister chan *Connection

	// Flag to determine if running or not
	isRunning bool
}

func (hub *Hub) Remove(conn *Connection) {
	delete(hub.connections, conn)
	close(conn.out)
}

func (hub *Hub) Run() {
	if hub.isRunning != true {
		hub.isRunning = true
		go hub.run()
	}
}

func (hub *Hub) run() {
	for {
		select {
		case conn := <-hub.register:
			hub.connections[conn] = true
		case conn := <-hub.unregister:
			hub.Remove(conn)
		case message := <-hub.broadcast:
			for conn := range hub.connections {
				select {
				case conn.out <- message:
				default: // default only triggered when sending failed, so get rid of problematic connection
					hub.Remove(conn)
					// go conn.CloseSocket() Shouldn't be necessary!
				}
			}
		}
	}
}

var hub = Hub{
	broadcast:   make(chan string),
	register:    make(chan *Connection),
	unregister:  make(chan *Connection),
	connections: make(map[*Connection]bool),
	isRunning:   false,
}
