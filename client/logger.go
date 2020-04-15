package client

import (
	"fmt"
	"time"
)

// NOTE: see https://github.com/go-log/log for the reasoning behind this minimal interface
// TODO: decide whether we want this minimal interface, or provide a debug, little bit more verbose, one?
// TODO: some way to indicate the loggin level we want and/or just info/debug/error?
type Logger interface {
	Log(v ...interface{})
	Logf(format string, v ...interface{})
}

func (c *Client) info(message string) {

	if c.logger == nil {
		return
	}

	c.logger.Log(fmt.Sprint(time.Now().UTC().Format("2006-01-02T15:04:05.999Z") + "  [INFO] " + message))
}

func (c *Client) debug(message string) {

	if c.logger == nil {
		return
	}

	c.logger.Log(fmt.Sprint(time.Now().UTC().Format("2006-01-02T15:04:05.999Z") + " [DEBUG] " + message))
}
