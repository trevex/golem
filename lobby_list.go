package golem

type lobbyList struct {
	lobbies map[*lobby]bool
	add     chan *lobby
	remove  chan *lobby
	clear   chan *Connection
	stop    chan bool
}

func newLobbyList() *lobbyList {
	return &lobbyList{
		lobbies: make(map[*lobby]bool),
		add:     make(chan *lobby),
		remove:  make(chan *lobby),
		clear:   make(chan *Connection),
		stop:    make(chan bool),
	}
}

func (ll *lobbyList) run() {
	for {
		select {
		case l := <-ll.add:
			ll.lobbies[l] = true
		case l := <-ll.remove:
			_, ok := ll.lobbies[l]
			if ok {
				delete(ll.lobbies, l)
			}
		case conn := <-ll.clear:
			for l := range ll.lobbies {
				l.unsubscribe <- conn
				delete(ll.lobbies, l)
			}
		case <-ll.stop:
			return
		}
	}
}
