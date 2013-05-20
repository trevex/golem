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
	"flag"
	"log"
)

/*
	should be rewritten and use a trick like this:
	http://play.golang.org/p/8o0WywyaDT
*/

var debug *bool = flag.Bool("debug", true, "enable debug logging") // currently true for development

const (
	delimiter = " "
	FATAL     = "[FATAL]"
	ERROR     = "[ERROR]"
	WARN      = "[WARN ]"
	DEBUG     = "[DEBUG]"
	INFO      = "[INFO ]"
)

func Debugf(format string, args ...interface{}) {
	if *debug {
		log.Printf(DEBUG+delimiter+format, args)
	}
}

func Errorf(format string, args ...interface{}) {
	log.Printf(ERROR+delimiter+format, args)
}
