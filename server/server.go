// Package server describes game servers that accept RCON connections.
package server

import (
	"errors"
	"fmt"
	"net"
	"strconv"
)

var (
	// ErrEmptyAddress is returned by [Server.Validate] when Address is empty.
	ErrEmptyAddress = errors.New("empty address")

	// ErrInvalidPort is returned by [Server.Validate] when Port is outside 1–65535.
	ErrInvalidPort = errors.New("invalid port")
)

// Key identifies a server by its address and port.
type Key struct {
	Address string
	Port    int
}

// String returns the key in "address:port" form.
func (k Key) String() string {
	return net.JoinHostPort(k.Address, strconv.Itoa(k.Port))
}

// Server holds the parameters needed to connect to a game server over RCON.
type Server struct {
	// Address is the host name or IP address of the server.
	Address string

	// Port is the RCON port, e.g. 25575 for Minecraft or 27015 for Source games.
	Port int

	// Password is the RCON password.
	Password string

	// Type selects the RCON protocol implementation.
	Type Type
}

// Key returns the key that identifies s.
func (s Server) Key() Key {
	return Key{Address: s.Address, Port: s.Port}
}

// New returns a Server with the given parameters.
func New(address string, port int, password string, serverType Type) Server {
	return Server{
		Address:  address,
		Port:     port,
		Password: password,
		Type:     serverType,
	}
}

// Validate returns [ErrEmptyAddress] or [ErrInvalidPort] if s cannot be connected to.
// It does not check Type; unsupported types are reported when the server is registered.
func (s Server) Validate() error {
	if s.Address == "" {
		return ErrEmptyAddress
	}
	if s.Port < 1 || s.Port > 65535 {
		return fmt.Errorf("%w: %d", ErrInvalidPort, s.Port)
	}
	return nil
}
