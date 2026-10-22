// Package rcon defines the interface shared by RCON client implementations.
package rcon

import "context"

// Client is a connection to a single server over RCON.
// Implementations must be safe for concurrent use.
type Client interface {
	// Connect opens the connection. Calling it is optional: Execute connects
	// on demand.
	Connect(ctx context.Context) error

	// Execute sends cmd to the server and returns its response.
	Execute(ctx context.Context, cmd string) (string, error)

	// Disconnect closes the current connection. A later Execute reconnects.
	Disconnect()

	// Close closes the connection permanently. After Close, Connect and
	// Execute return [ErrClosed].
	Close() error
}
