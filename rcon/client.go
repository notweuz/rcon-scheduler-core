package rcon

import "context"

type Client interface {
	Connect(ctx context.Context) error
	Execute(ctx context.Context, cmd string) (string, error)
	Disconnect()
	Close() error
}