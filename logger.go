package gohelix

import (
	"io/ioutil"
	"log"
)

// Logger is the instance of a StdLogger interface that Sarama writes connection
// management events to. By default it is set to discard all log messages via ioutil.Discard,
// but you can set it to redirect wherever you want.
var Logger StdLogger = log.New(ioutil.Discard, "[gohelix] ", log.LstdFlags)

// StdLogger is used to log error messages.
type StdLogger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}
