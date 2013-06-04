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
// 1. Event name needs to be splitted/extracted from incoming data.
// 2. After event is known, the second part of the data needs to be unmarshalled into the associated type.
// For emitting data the process is reversed:
// 1. Marshal data into byte array.
// 2. Pack event name into byte array.
type Protocol interface {
	Unmarshal([]byte, interface{}) error
	Marshal(interface{}) ([]byte, error)
	Unpack([]byte) (string, []byte, error)
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

type DefaultJSONProtocol struct{}

func (_ *DefaultJSONProtocol) Unmarshal(data []byte, structPtr interface{}) error {
	return json.Unmarshal(data, structPtr)
}

func (_ *DefaultJSONProtocol) Marshal(structPtr interface{}) ([]byte, error) {
	return json.Marshal(structPtr)
}

func (_ *DefaultJSONProtocol) Unpack(data []byte) (string, []byte, error) {
	result := strings.SplitN(string(data), protocolSeperator, 2)
	if len(result) != 2 {
		return "", nil, errors.New("Unable to extract event name from data.")
	}
	return result[0], []byte(result[1]), nil
}

func (_ *DefaultJSONProtocol) Pack(name string, data []byte) ([]byte, error) {
	result := []byte(name + protocolSeperator)
	result = append(result, data...)
	return result, nil
}
