package client

import "errors"

var (
	ErrNotConnected = errors.New("not connected")
)

const Attempts = 5

type RconClient interface {
	Connect()
	Execute(cmd string) (string, error)
	Disconnect()
}
