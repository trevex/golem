package golem

type Hub struct {
	// Connection Manager
	connMngr *connectionManager

	// Flag to determine if running or not
	isRunning bool
}

func (hub *Hub) run() {
	if hub.isRunning != true {
		hub.isRunning = true
		go hub.connMngr.run()
	}
}

var hub = Hub{
	connMngr:  newConnectionManager(),
	isRunning: false,
}

func GetHub() *Hub {
	return &hub
}
