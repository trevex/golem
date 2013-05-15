package golem

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"fmt"
	"net/http"
)

type DataType map[string]interface{}
type CallbackType func(*Connection, *DataType)
type CallbackMap map[string]CallbackType

type Protocol struct {
	CallbackName string   `json:"n"`
	Data         DataType `json:"d"`
}

type Router struct {
	callbacks CallbackMap
}

func NewRouter() *Router {
	hub.Run()
	return &Router{
		callbacks: make(CallbackMap),
	}
}

func (router *Router) On(name string, cb CallbackType) {
	router.callbacks[name] = cb
}

func (router *Router) Handler() http.Handler {
	return websocket.Handler(CreateHandler(router))
}

func (router *Router) Parse(conn *Connection, message string) {
	var decoded Protocol
	err := json.Unmarshal([]byte(message), &decoded)
	if err == nil {
		if cb, ok := router.callbacks[decoded.CallbackName]; ok != false {
			cb(conn, &decoded.Data)
		}
	} else {
		fmt.Println("[JSON-FAIL]", message) // TODO: Proper debug output!
	}
	defer recover()
}

func CreateHandler(router *Router) func(*websocket.Conn) {
	return func(socket *websocket.Conn) {
		conn := &Connection{
			socket: socket,
			router: router,
			out:    make(chan string, outChannelSize),
		}
		hub.register <- conn
		defer func() { hub.unregister <- conn }()
		go conn.Writer()
		conn.Reader()
	}
}
