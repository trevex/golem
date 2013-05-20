/*

	golem - lightweight Go WebSocket-framework
    Copyright (C) 2013  Niklas Voss

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <http://www.gnu.org/licenses/>.

*/

package golem

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/go-websocket/websocket"
	"log"
	"net/http"
	"reflect"
	"strings"
)

const (
	protocolSeperator = " "
)

type Router struct {
	callbacks map[string]func(*Connection, []byte)

	closeCallback func(*Connection)
}

func NewRouter() *Router {
	hub.run()
	return &Router{
		callbacks:     make(map[string]func(*Connection, []byte)),
		closeCallback: func(*Connection) {},
	}
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

		conn := newConnection(socket, router)

		hub.register <- conn
		go conn.writePump()
		conn.readPump()
	}
}

func (router *Router) On(name string, callback interface{}) {

	callbackDataType := reflect.TypeOf(callback).In(1)

	// If function accepts byte arrays, use NO parser
	if reflect.TypeOf([]byte{}) == callbackDataType {
		router.callbacks[name] = callback.(func(*Connection, []byte))
		return
	}

	// Needed by custom and json parsers
	callbackValue := reflect.ValueOf(callback)

	// If parser is available for this type, use it
	if parser, ok := parserMap[callbackDataType]; ok {
		parserThenCallback := func(conn *Connection, data []byte) {
			if result := parser.Call([]reflect.Value{reflect.ValueOf(data)}); result[1].Bool() {
				args := []reflect.Value{reflect.ValueOf(conn), result[0]}
				callbackValue.Call(args)
			}
		}
		router.callbacks[name] = parserThenCallback
		return
	}

	// Else interpret data as JSON and try to unmarshal it into requested type
	callbackDataElem := callbackDataType.Elem()
	unmarshalThenCallback := func(conn *Connection, data []byte) {
		result := reflect.New(callbackDataElem)

		err := json.Unmarshal(data, result.Interface())
		if err == nil {
			args := []reflect.Value{reflect.ValueOf(conn), result}
			callbackValue.Call(args)
		} else {
			fmt.Println("[JSON-FORWARD]", data, err) // TODO: Proper debug output!
		}
	}
	router.callbacks[name] = unmarshalThenCallback
}

func (router *Router) parse(conn *Connection, rawdata []byte) {
	rawstring := string(rawdata)
	data := strings.SplitN(rawstring, protocolSeperator, 2)
	if callback, ok := router.callbacks[data[0]]; ok {
		callback(conn, []byte(data[1]))
	}

	defer recover()
}

func (router *Router) OnClose(callback func(*Connection)) {
	router.closeCallback = callback
}
