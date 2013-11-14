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
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"reflect"
)

// Router handles multiplexing of incoming messenges by typenames/events.
// Initially a router uses heartbeats and the default protocol.
type Router struct {
	// Map of callbacks for event types.
	callbacks map[string]func(*Connection, interface{})
	// Protocol extensions
	extensions map[reflect.Type]reflect.Value
	// Function being called if connection is closed.
	closeFunc func(*Connection)
	// Function called after handshake when a WebSocket connection
	// was succesfully established.
	connectionFunc func(*Connection, *http.Request)
	// Function verifying handshake.
	handshakeFunc func(http.ResponseWriter, *http.Request) bool
	// Active protocol
	protocol Protocol
	// Flag to enable or disable heartbeats
	useHeartbeats bool
	//
	connExtensionConstructor reflect.Value
	// If set, the values the Origin header will be checked against and access is only allowed
	// on a match; otherwise no Origin checking is performed. *This overrides the
	// Access-Control-Allow-Origin header!*
	Origins []string
}

// NewRouter intialises a new instance and returns the pointer.
func NewRouter() *Router {
	// Tries to run hub, if already running nothing will happen.
	hub.run()
	// Returns pointer to instance.
	return &Router{
		callbacks:                make(map[string]func(*Connection, interface{})),
		extensions:               make(map[reflect.Type]reflect.Value),
		closeFunc:                func(*Connection) {}, // Empty placeholder close function.
		connectionFunc:           func(*Connection, *http.Request) {},
		handshakeFunc:            func(http.ResponseWriter, *http.Request) bool { return true }, // Handshake always allowed.
		protocol:                 initialProtocol,
		useHeartbeats:            true,
		connExtensionConstructor: defaultConnectionExtension,
		Origins:                  make([]string, 0),
	}
}

// Handler creates a handler function for this router, that can be used with the
// http-package to handle WebSocket-Connections.
func (router *Router) Handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if method used was GET.
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", 405)
			return
		}

		// Disallow cross-origin connections.
		if len(router.Origins) > 0 {
			originFound := false
			for _, origin := range router.Origins {
				if r.Header.Get("Origin") == origin {
					originFound = true
					break
				}
			}
			if !originFound {
				http.Error(w, "Origin not allowed", 403)
				return
			}
		} else {
			if len(r.Header.Get("Access-Control-Allow-Origin")) > 0 {
				allowedOrigin := r.Header.Get("Access-Control-Allow-Origin")
				if allowedOrigin != "*" {
					if r.URL.Scheme+"://"+r.Host != allowedOrigin {
						http.Error(w, "Origin not allwed", 403)
						return
					}
				}
			}
		}

		// Check if handshake callback verifies upgrade.
		if !router.handshakeFunc(w, r) {
			http.Error(w, "Authorization failed", 403)
			return
		}

		// Upgrade websocket connection.
		protocols := websocket.Subprotocols(r)
		var responseHeader http.Header = nil
		if len(protocols) > 0 {
			responseHeader = http.Header{"Sec-Websocket-Protocol": {protocols[0]}}
		}
		socket, err := websocket.Upgrade(w, r, responseHeader, 1024, 1024)
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
		//
		if router.connExtensionConstructor.IsValid() {
			conn.extend(router.connExtensionConstructor.Call([]reflect.Value{reflect.ValueOf(conn)})[0].Interface())
		}

		// Connection established with possible extension, so callback
		router.connectionFunc(conn, r)

		// And start reading and writing routines.
		conn.run()
	}
}

// The On-function adds callbacks by name of the event, that should be handled.
// For type T the callback would be of type:
//     func (*golem.Connection, *T)
// Type T can be any type. By default golem tries to unmarshal json into the
// specified type. If a custom protocol is used, it will be used instead to process the data.
// If type T is registered to use a protocol extension, it will be used instead.
// If type T is interface{} the interstage data of the active protocol will be directly forwarded!
// (Note: the golem wiki has a whole page about this function)
func (router *Router) On(name string, callback interface{}) {

	callbackValue := reflect.ValueOf(callback)
	callbackType := reflect.TypeOf(callback)
	if router.connExtensionConstructor.IsValid() {
		extType := router.connExtensionConstructor.Type().Out(0)
		if callbackType.In(0) == extType {
			// EXTENSION TYPE

			// NO DATA
			if callbackType.NumIn() == 1 {
				router.callbacks[name] = func(conn *Connection, data interface{}) {
					args := []reflect.Value{reflect.ValueOf(conn.extension)}
					callbackValue.Call(args)
				}
				return
			}

			// INTERFACE
			if callbackType.In(1).Kind() == reflect.Interface {
				router.callbacks[name] = func(conn *Connection, data interface{}) {
					args := []reflect.Value{reflect.ValueOf(conn.extension), reflect.ValueOf(data)}
					callbackValue.Call(args)
				}
				return
			}

			// PROTOCOL EXTENSION
			if parser, ok := router.extensions[callbackType.In(1)]; ok {
				router.callbacks[name] = func(conn *Connection, data interface{}) {
					if result := parser.Call([]reflect.Value{reflect.ValueOf(data)}); result[1].Bool() {
						args := []reflect.Value{reflect.ValueOf(conn.extension), result[0]}
						callbackValue.Call(args)
					}
				}
				return
			}

			// PROTOCOL
			callbackDataElem := callbackType.In(1).Elem()
			router.callbacks[name] = func(conn *Connection, data interface{}) {
				result := reflect.New(callbackDataElem)

				err := router.protocol.Unmarshal(data, result.Interface())
				if err == nil {
					args := []reflect.Value{reflect.ValueOf(conn.extension), result}
					callbackValue.Call(args)
				} else {
					// TODO: Proper debug output!
				}
			}
			return
		}
	} else {
		// DEFAULT TYPE

		// NO DATA
		if reflect.TypeOf(callback).NumIn() == 1 {
			router.callbacks[name] = func(conn *Connection, data interface{}) {
				callback.(func(*Connection))(conn)
			}
			return
		}

		// INTERFACE
		if cb, ok := callback.(func(*Connection, interface{})); ok {
			router.callbacks[name] = cb
			return
		}

		// PROTOCOL EXTENSION
		if parser, ok := router.extensions[callbackType.In(1)]; ok {
			router.callbacks[name] = func(conn *Connection, data interface{}) {
				if result := parser.Call([]reflect.Value{reflect.ValueOf(data)}); result[1].Bool() {
					args := []reflect.Value{reflect.ValueOf(conn), result[0]}
					callbackValue.Call(args)
				}
			}
			return
		}

		// PROTOCOL
		callbackDataElem := callbackType.In(1).Elem()
		router.callbacks[name] = func(conn *Connection, data interface{}) {
			result := reflect.New(callbackDataElem)

			err := router.protocol.Unmarshal(data, result.Interface())
			if err == nil {
				args := []reflect.Value{reflect.ValueOf(conn), result}
				callbackValue.Call(args)
			} else {
				// TODO: Proper debug output!
			}
		}
		return
	}
}

// Unpacks incoming data and forwards it to callback.
func (router *Router) processMessage(conn *Connection, in []byte) {
	if name, data, err := router.protocol.Unpack(in); err == nil {
		if callback, ok := router.callbacks[name]; ok {
			callback(conn, data)
		}
	} // TODO: else error logging?

	defer recover()
}

// OnClose sets the callback, that is called when the connection is closed.
// It accept function of the type func(*Connection) by default or functions
// taking extended connection types if previously registered.
func (router *Router) OnClose(callback interface{}) error { //func(*Connection)) {
	if cb, ok := callback.(func(*Connection)); ok {
		router.closeFunc = cb
	} else {
		if router.connExtensionConstructor.IsValid() {
			callbackValue := reflect.ValueOf(callback)
			extType := router.connExtensionConstructor.Type().Out(0)
			if reflect.TypeOf(callback).In(0) == extType {
				router.closeFunc = func(conn *Connection) {
					callbackValue.Call([]reflect.Value{reflect.ValueOf(conn.extension)})
				}
			} else {
				return errors.New("OnClose cannot accept a callback of the type " + reflect.TypeOf(callback).String() + ".")
			}
		} else {
			return errors.New("OnClose can only accept functions of the type func(*Connection), if no extension is registered.")
		}
	}
	return nil
}

// OnConnection sets the callback, that is called when a websocket connection
// was successfully established, it is therefore called after the handshake.
// It accept function of the type func(*Connection, *http.Request) by default or functions
// taking extended connection types if previously registered.
// The http.Request object can be used for connection metadata information.
func (router *Router) OnConnect(callback interface{}) error { //func(*Connection)) {
	if cb, ok := callback.(func(*Connection, *http.Request)); ok {
		router.connectionFunc = cb
	} else {
		if router.connExtensionConstructor.IsValid() {
			callbackValue := reflect.ValueOf(callback)
			extType := router.connExtensionConstructor.Type().Out(0)
			if reflect.TypeOf(callback).In(0) == extType {
				router.connectionFunc = func(conn *Connection, hr *http.Request) {
					callbackValue.Call([]reflect.Value{reflect.ValueOf(conn.extension), reflect.ValueOf(hr)})
				}
			} else {
				return errors.New("OnConnection cannot accept a callback of the type " + reflect.TypeOf(callback).String() + ".")
			}
		} else {
			return errors.New("OnConnection can only accept functions of the type func(*Connection), if no extension is registered.")
		}
	}
	return nil
}

// OnHandshake sets the callback for handshake verfication.
// If the handshake function returns false the request will not be upgraded.
// The http.Request object will be passed into OnConnect as well.
func (router *Router) OnHandshake(callback func(http.ResponseWriter, *http.Request) bool) {
	router.handshakeFunc = callback
}

// The AddProtocolExtension-function allows adding of custom parsers for custom types. For any Type T
// the parser function would look like this:
//      func (interface{}) (T, bool)
// The interface's type is depending on the interstage product of the active protocol, by default
// for the JSON-based protocol it is []byte and therefore the function could be simplified to:
//      func ([]byte) (T, bool)
// Or in general if P is the interstage product:
//      func (*P) (T, bool)
// The boolean return value is necessary to verify if parsing was successful.
// All On-handling function accepting T as input data will now automatically use the custom
// extension. For an example see the example_data.go file in the example repository.
func (router *Router) AddProtocolExtension(extensionFunc interface{}) error {
	extensionValue := reflect.ValueOf(extensionFunc)
	extensionType := extensionValue.Type()

	if extensionType.NumIn() != 1 {
		return errors.New("Cannot add function(" + extensionType.String() + ") as parser: To many arguments!")
	}
	if extensionType.NumOut() != 2 {
		return errors.New("Cannot add function(" + extensionType.String() + ") as parser: Wrong number of return values!")
	}
	if extensionType.Out(1).Kind() != reflect.Bool {
		return errors.New("Cannot add function(" + extensionType.String() + ") as parser: Second return value is not Bool!")
	}

	router.extensions[extensionType.Out(0)] = extensionValue

	return nil
}

// SetConnectionExtension sets the extension for this router. A connection extension is an extended connection structure, that afterwards can be
// used by On-handlers as well (use-cases: add persistent storage to connection, additional methods et cetera).
// The SetConnectionExtension function takes the constructor of the custom format to be able to use and create it on
// connection to the router.
// For type E the constructor needs to fulfil the following requirements:
//     func NewE(conn *Connection) *E
// Afterwards On-handler can us this extended type:
//     router.On(func (extendedConn E, data Datatype) { ... })
// For an example have a look at the example repository and have a look at the 'example_connection_extension.go'.
func (router *Router) SetConnectionExtension(constructor interface{}) {
	router.connExtensionConstructor = reflect.ValueOf(constructor)
}

// SetProtocol sets the protocol of the router to the supplied implementation of the Protocol interface.
func (router *Router) SetProtocol(protocol Protocol) {
	router.protocol = protocol
}

//

// SetHeartbeat activates or deactivates the heartbeat depending on the flag parameter. By default heartbeats are activated.
func (router *Router) SetHeartbeat(flag bool) {
	router.useHeartbeats = flag
}
