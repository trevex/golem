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
	"strings"
)

// Marshal any data into json and prepend the name of the event.
func pack(name string, data interface{}) ([]byte, bool) {
	result := []byte(name + protocolSeperator)
	b, err := json.Marshal(data)
	if err != nil {
		return result, false
	}
	result = append(result, b...)
	return result, true
}

// Split the name of the event to get the raw data.
func unpack(in []byte) (string, []byte) {
	data := strings.SplitN(string(in), protocolSeperator, 2)
	return data[0], []byte(data[1])
}
