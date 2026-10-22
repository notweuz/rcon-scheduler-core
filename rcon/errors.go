package rcon

import "errors"

// ErrClosed is returned by a [Client] after Close has been called.
var ErrClosed = errors.New("rcon client closed")
