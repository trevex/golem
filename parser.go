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
