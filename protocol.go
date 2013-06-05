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
	"encoding/json"
	"errors"
	"strings"
)

const (
	protocolSeperator = " "
	BinaryMode        = 1
	TextMode          = 2
)

var (
	initialProtocol Protocol = Protocol(&DefaultJSONProtocol{})
)

// Protocol-interface provides the required methods necessary for any
// protocol, that should be used with golem, to implement.
// The evented system of golem needs several steps to process incoming data:
//  1. Unpack
//  2. Unmarshal
// For emitting data the process is reversed, but merged in a single function,
// because evaluation the desired unmarshaled type is not necessary:
//  1. MarshalAndPack
// The GetReadMode and GetWriteMode functions define what kind of Operation-Code
// the sockets will receive.
type Protocol interface {
	// Unpack splits/extracts event name from incoming data.
	Unpack([]byte) (string, interface{}, error)
	// Unmarshals leftover data into associated type of callback.
	Unmarshal(interface{}, interface{}) error
	// Marshal and pack data into byte array
	MarshalAndPack(string, interface{}) ([]byte, error)
	// Returns read mode, that should be used for this protocol.
	GetReadMode() int
	// Returns write mode, that should be used for this protocol
	GetWriteMode() int
}

// SetInitialProtocol sets the initial protocol for router creation. Every router
// created after changing the initial protocol will use the new protocol by default.
func SetInitialProtocol(protocol Protocol) {
	initialProtocol = protocol
}

// DefaultJSONProtocol is the initial protocol used by golem. It implements the
// Protocol-Interface.
// (Note: there is an article about this simple protocol in golem's wiki)
type DefaultJSONProtocol struct{}

// Unpack splits the event name from the incoming message.
func (_ *DefaultJSONProtocol) Unpack(data []byte) (string, interface{}, error) {
	result := strings.SplitN(string(data), protocolSeperator, 2)
	if len(result) != 2 {
		return "", nil, errors.New("Unable to extract event name from data.")
	}
	return result[0], []byte(result[1]), nil
}

// Unmarshals data into requested structure. If not successful the function return an error.
func (_ *DefaultJSONProtocol) Unmarshal(data interface{}, typePtr interface{}) error {
	return json.Unmarshal(data.([]byte), typePtr)
}

// Marshals structure into JSON and pack event name in aswell. If not successful second return value is an error.
func (_ *DefaultJSONProtocol) MarshalAndPack(name string, structPtr interface{}) ([]byte, error) {
	if data, err := json.Marshal(structPtr); err != nil {
		result := []byte(name + protocolSeperator)
		result = append(result, data...)
		return result, nil
	} else {
		return nil, err
	}
}

// Return TextMode because JSON is transmitted using the text mode of WebSockets
func (_ *DefaultJSONProtocol) GetReadMode() int {
	return TextMode
}

// Return TextMode because JSON is transmitted using the text mode of WebSockets
func (_ *DefaultJSONProtocol) GetWriteMode() int {
	return TextMode
}
