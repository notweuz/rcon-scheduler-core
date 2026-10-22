package server

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyAddress = errors.New("empty address")
	ErrInvalidPort  = errors.New("invalid port")
)

type Key struct {
	Address string
	Port    int
}

func (k Key) String() string {
	return fmt.Sprintf("%v:%d", k.Address, k.Port)
}

type Server struct {
	Address  string
	Port     int
	Password string
	Type     Type
}

func (s Server) Key() Key {
	return Key{Address: s.Address, Port: s.Port}
}

func New(address string, port int, password string, serverType Type) Server {
	return Server{
		Address:  address,
		Port:     port,
		Password: password,
		Type:     serverType,
	}
}

func (s Server) Validate() error {
	if s.Address == "" {
		return ErrEmptyAddress
	}
	if s.Port < 1 || s.Port > 65535 {
		return fmt.Errorf("%w: %d", ErrInvalidPort, s.Port)
	}
	return nil
}
