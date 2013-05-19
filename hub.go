package golem

type Hub struct {
	// Connection Manager
	connMgr *connectionManager

	// Lobby Manager
	lobbyMgr *lobbyManager

	// Flag to determine if running or not
	isRunning bool
}

func (hub *Hub) run() {
	if hub.isRunning != true {
		hub.isRunning = true
		go hub.connMgr.run()
		go hub.lobbyMgr.run()
	}
}

var hub = Hub{
	connMgr:   newConnectionManager(),
	lobbyMgr:  newLobbyManager(),
	isRunning: false,
}

func GetHub() *Hub {
	return &hub
}
