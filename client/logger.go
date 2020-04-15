package client

// NOTE: see https://github.com/go-log/log for the reasoning behind this minimal interface
// TODO: decide whether we want this minimal interface, or provide a debug, little bit more verbose, one?
type Logger interface {
	Log(v ...interface{})
	Logf(format string, v ...interface{})
}

func (c *Client) info(message string) {

	if c.logger == nil {
		return
	}

	c.logger.Log(message)
}
