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

var (
	initialProtocol Protocol = Protocol(&DefaultJSONProtocol{})
)

// Protocol-interface provides the required methods necessary for any
// protocol, that should be used with golem, to implement.
// The evented system of golem needs several steps to process incoming data:
//  1. Unpack
//  2. Unmarshal
// For emitting data the process is reversed:
//  1. Marshal
//  2. Pack
type Protocol interface {
	// Unpack splits/extracts event name from incoming data.
	Unpack([]byte) (string, []byte, error)
	// Unmarshals leftover data into associated type of callback.
	Unmarshal([]byte, interface{}) error
	// Marshal data into byte array
	Marshal(interface{}) ([]byte, error)
	// Pack event name into byte array.
	Pack(string, []byte) ([]byte, error)
}

// SetInitialProtocol sets the initial protocol for router creation. Every router
// created after changing the initial protocol will use the new protocol by default.
func SetInitialProtocol(protocol Protocol) {
	initialProtocol = protocol
}

const (
	protocolSeperator = " "
)

// DefaultJSONProtocol is the initial protocol used by golem. It implements the
// Protocol-Interface.
// (Note: there is an article about this simple protocol in golem's wiki)
type DefaultJSONProtocol struct{}

// Unpack splits the event name from the incoming message.
func (_ *DefaultJSONProtocol) Unpack(data []byte) (string, []byte, error) {
	result := strings.SplitN(string(data), protocolSeperator, 2)
	if len(result) != 2 {
		return "", nil, errors.New("Unable to extract event name from data.")
	}
	return result[0], []byte(result[1]), nil
}

// Unmarshals data into requested structure. If not successful the function return an error.
func (_ *DefaultJSONProtocol) Unmarshal(data []byte, structPtr interface{}) error {
	return json.Unmarshal(data, structPtr)
}

// Marshals structure into JSON. If not successful second return value is an error.
func (_ *DefaultJSONProtocol) Marshal(structPtr interface{}) ([]byte, error) {
	return json.Marshal(structPtr)
}

// Adds the event name to the message.
func (_ *DefaultJSONProtocol) Pack(name string, data []byte) ([]byte, error) {
	result := []byte(name + protocolSeperator)
	result = append(result, data...)
	return result, nil
}
