/*

   Copyright 2013 Niklas Voss

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

*/

package golem

import (
	"github.com/garyburd/go-websocket/websocket"
	"log"
	"net/http"
	"reflect"
)

// Router handle multiplexing of incoming messenges by typename/event.
type Router struct {
	// Map of callbacks for event types.
	callbacks map[string]func(*Connection, []byte)
	// Function being called if connection is closed.
	closeCallback func(*Connection)
	// Function verifying handshake.
	handshakeCallback func(http.ResponseWriter, *http.Request) bool
	//
	protocol Protocol
}

// Returns new router instance.
func NewRouter() *Router {
	// Tries to run hub, if already running nothing will happen.
	hub.run()
	// Returns pointer to instance.
	return &Router{
		callbacks:         make(map[string]func(*Connection, []byte)),
		closeCallback:     func(*Connection) {},                                          // Empty placeholder close function.
		handshakeCallback: func(http.ResponseWriter, *http.Request) bool { return true }, // Handshake always allowed.
		protocol:          initialProtocol,
	}
}

// Creates handler function for this router.
func (router *Router) Handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if method used was GET.
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", 405)
			return
		}
		// Disallow cross-origin connections.
		if r.Header.Get("Origin") != "http://"+r.Host {
			http.Error(w, "Origin not allowed", 403)
			return
		}
		// Check if handshake callback verifies upgrade.
		if !router.handshakeCallback(w, r) {
			http.Error(w, "Authorization failed", 403)
			return
		}
		// Upgrade websocket connection.
		socket, err := websocket.Upgrade(w, r.Header, nil, 1024, 1024)
		// Check if handshake was successful
		if _, ok := err.(websocket.HandshakeError); ok {
			http.Error(w, "Not a websocket handshake", 400)
			return
		} else if err != nil {
			log.Println(err)
			return
		}

		// Create the connection.
		conn := newConnection(socket, router)
		// And start reading and writing routines.
		conn.run()
	}
}

// The On-function adds callbacks by name of the event, that should be handled.
// For type T the callback would be of type:
//     func (*golem.Connection, *T)
// Type T can be any type. By default golem tries to unmarshal json into the
// specified type. If a custom parser is known for the specified type, it will be
// used instead.
// If type T is []byte the incoming data will be directly forwarded!
func (router *Router) On(name string, callback interface{}) {

	// If callback function doesn't exept data
	if reflect.TypeOf(callback).NumIn() == 1 {
		router.callbacks[name] = func(conn *Connection, data []byte) {
			callback.(func(*Connection))(conn)
		}
		return
	}

	// Type of callback function
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

		err := router.protocol.Unmarshal(data, result.Interface())
		if err == nil {
			args := []reflect.Value{reflect.ValueOf(conn), result}
			callbackValue.Call(args)
		} else {
			// TODO: Proper debug output!
		}
	}
	router.callbacks[name] = unmarshalThenCallback
}

// Unpacks incoming data and forwards it to callback.
func (router *Router) parse(conn *Connection, in []byte) {
	if name, data, err := router.protocol.Unpack(in); err == nil {
		if callback, ok := router.callbacks[name]; ok {
			callback(conn, data)
		}
	} // TODO: else error logging?

	defer recover()
}

// Set the callback for connection closes.
func (router *Router) OnClose(callback func(*Connection)) {
	router.closeCallback = callback
}

// Set the callback for handshake verfication.
func (router *Router) OnHandshake(callback func(http.ResponseWriter, *http.Request) bool) {
	router.handshakeCallback = callback
}

func (router *Router) SetProtocol(protocol Protocol) {
	router.protocol = protocol
}

func (router *Router) prepareDataForEmit(name string, data interface{}) ([]byte, error) {
	if data, err := router.protocol.Marshal(data); err == nil {
		return router.protocol.Pack(name, data)
	} else {
		return nil, err
	}
}
