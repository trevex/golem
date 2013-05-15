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
