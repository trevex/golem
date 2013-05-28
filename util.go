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
	"strings"
)

const (
	protocolSeperator = " "
)

func pack(name string, data interface{}) ([]byte, bool) {
	result := []byte(name + protocolSeperator)
	b, err := json.Marshal(data)
	if err != nil {
		return result, false
	}
	result = append(result, b...)
	return result, true
}

func unpack(in []byte) (string, []byte) {
	data := strings.SplitN(string(in), protocolSeperator, 2)
	return data[0], []byte(data[1])
}
