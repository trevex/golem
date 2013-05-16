package golem

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/go-websocket/websocket"
	"log"
	"net/http"
	"reflect"
)

type DataType *json.RawMessage
type CallbackType func(*Connection, DataType)
type CallbackMap map[string]CallbackType

type Protocol map[string]DataType

func (p Protocol) GetName() (string, bool) {
	var name string
	err := json.Unmarshal(*p["n"], &name)
	if err == nil {
		return name, true
	} else {
		return name, false
	}
}

func (p Protocol) GetData() DataType {
	return p["d"]
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

func (router *Router) On(name string, cb interface{}) {
	cbValue := reflect.ValueOf(cb)
	cbDataType := reflect.TypeOf(cb).In(1)
	pre := func(conn *Connection, data DataType) {
		decoded := reflect.New(cbDataType)
		err := json.Unmarshal(*data, &decoded)
		if err == nil {
			args := []reflect.Value{reflect.ValueOf(conn), reflect.ValueOf(decoded)}
			cbValue.Call(args)
		} else {
			fmt.Println("[JSON-FORWARD]", data, err) // TODO: Proper debug output!
		}
	}
	router.callbacks[name] = pre
}

func (router *Router) Parse(conn *Connection, message []byte) {
	var decoded Protocol
	err := json.Unmarshal(message, &decoded)
	if err == nil {
		name, ok := decoded.GetName()
		if ok {
			if cb, ok := router.callbacks[name]; ok != false {
				cb(conn, decoded.GetData())
			}
		} else {
			fmt.Println("[JSON-NAME]", string(message), err) // TODO: Proper debug output!
		}
	} else {
		fmt.Println("[JSON-INTERPRET]", string(message), err) // TODO: Proper debug output!
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
