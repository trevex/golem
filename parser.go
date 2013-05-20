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
	"fmt"
	"reflect"
)

var parserMap map[reflect.Type]reflect.Value = make(map[reflect.Type]reflect.Value)

func AddParser(parserFn interface{}) {
	parserValue := reflect.ValueOf(parserFn)
	parserType := parserValue.Type()

	if parserType.NumIn() != 1 {
		fmt.Println("Cannot add function(", parserType, ") as parser: To many arguments!")
		return
	}
	if parserType.NumOut() != 2 {
		fmt.Println("Cannot add function(", parserType, ") as parser: Wrong number of return values!")
		return
	}
	if parserType.Out(1).Kind() != reflect.Bool {
		fmt.Println("Cannot add function(", parserType, ") as parser: Second return value is not Bool!")
		return
	}

	parserMap[parserType.Out(0)] = parserValue
}
