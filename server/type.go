package server

// Type is an RCON protocol implementation.
type Type string

const (
	// TypeSource is the Source RCON protocol, also used by Minecraft.
	TypeSource Type = "source"

	// TypeWebsocket is WebSocket RCON. Not implemented yet.
	TypeWebsocket Type = "websocket"

	// TypeTelnet is Telnet RCON. Not implemented yet.
	TypeTelnet Type = "telnet"
)
