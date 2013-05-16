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
