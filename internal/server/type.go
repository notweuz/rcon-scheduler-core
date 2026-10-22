package server

type Type string

const (
	TypeSource    Type = "source"
	TypeWebsocket Type = "websocket"
	TypeTelnet    Type = "telnet"
)