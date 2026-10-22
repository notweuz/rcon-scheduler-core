package rcon

type Client interface {
	Connect() error
	Execute(cmd string) (string, error)
	Disconnect()
}