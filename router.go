package golem

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/go-websocket/websocket"
	"log"
	"net/http"
	"reflect"
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

func (router *Router) On(name string, fptr CallbackType) {
	fn := reflect.ValueOf(fptr)
	fmt.Println(fn.Type())
	router.callbacks[name] = fptr
}

func (router *Router) Parse(conn *Connection, message []byte) {
	var decoded Protocol
	err := json.Unmarshal(message, &decoded)
	if err == nil {
		if cb, ok := router.callbacks[decoded.CallbackName]; ok != false {
			cb(conn, &decoded.Data)
		}
	} else {
		fmt.Println("[JSON-FAIL]", message) // TODO: Proper debug output!
	}
	defer recover()
}

func (router *Router) Handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "GET" {
			http.Error(w, "Method not allowed", 405)
			return
		}

		if r.Header.Get("Origin") != "http://"+r.Host {
			http.Error(w, "Origin not allowed", 403)
			return
		}

		socket, err := websocket.Upgrade(w, r.Header, nil, 1024, 1024)

		if _, ok := err.(websocket.HandshakeError); ok {
			http.Error(w, "Not a websocket handshake", 400)
			return
		} else if err != nil {
			log.Println(err)
			return
		}

		conn := &Connection{
			socket: socket,
			router: router,
			out:    make(chan []byte, outChannelSize),
		}

		hub.register <- conn
		go conn.writePump()
		conn.readPump()
	}
}
