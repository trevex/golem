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
	"errors"
	"reflect"
)

// Package-intern map of parser callbacks
var parserMap map[reflect.Type]reflect.Value = make(map[reflect.Type]reflect.Value)

// The AddParser-function allows adding of custom parsers for custom types. For any Type T
// the parser function would look like this:
//      func ([]byte) (T, bool)
// If the parser function does not follow this guideline an error is returned. The boolean is
// necessary to verify if parsing was successful.
// All On-handling function accepting T as input data will now automatically use the custom
// parser, e.g:
//      func (conn *golem.Connection, data string)
func AddParser(parserFn interface{}) error {
	parserValue := reflect.ValueOf(parserFn)
	parserType := parserValue.Type()

	if parserType.NumIn() != 1 {
		return errors.New("Cannot add function(" + parserType.String() + ") as parser: To many arguments!")
	}
	if parserType.NumOut() != 2 {
		return errors.New("Cannot add function(" + parserType.String() + ") as parser: Wrong number of return values!")
	}
	if parserType.Out(1).Kind() != reflect.Bool {
		return errors.New("Cannot add function(" + parserType.String() + ") as parser: Second return value is not Bool!")
	}

	parserMap[parserType.Out(0)] = parserValue

	return nil
}
