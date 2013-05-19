package golem

type lobbyRequest struct {
	name string
	conn *Connection
}

type lobbyManager struct {
	lobbies    map[string]*lobby
	register   chan *lobbyRequest
	unregister chan *lobbyRequest
	remove     chan *lobby
}

func newLobbyManager() *lobbyManager {
	return &lobbyManager{
		lobbies:    make(map[string]*lobby),
		register:   make(chan *lobbyRequest),
		unregister: make(chan *lobbyRequest),
		remove:     make(chan *lobby),
	}
}

func (lm *lobbyManager) run() {
	for {
		select {
		case req := <-lm.register:
			l, ok := lm.lobbies[req.name]
			if !ok {
				l := newLobby(lm)
				lm.lobbies[req.name] = l
				l.subscribe <- req.conn
				go l.run()
			} else {
				l.subscribe <- req.conn
			}
		}
	}
}
