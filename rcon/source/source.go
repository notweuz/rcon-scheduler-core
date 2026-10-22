package source

import (
	"errors"
	"fmt"
	"time"

	iRcon "github.com/notweuz/rcon-scheduler-core/rcon"
	"github.com/notweuz/rcon-scheduler-core/server"

	"github.com/gorcon/rcon"
	"github.com/rs/zerolog/log"
)

// TODO: apparently this needs mutex so it won't read other jobs while it's reading the current one
type Source struct {
	conn *rcon.Conn
	srv  server.Server
}

const attempts = 5

func NewSource(srv server.Server) iRcon.Client {
	return &Source{
		srv: srv,
	}
}

func (s *Source) Connect() error {
	address := fmt.Sprintf("%v:%d", s.srv.Address, s.srv.Port)
	var err error
	for i := 0; i < attempts; i++ {
		conn, err1 := rcon.Dial(address, s.srv.Password, rcon.SetDialTimeout(5*time.Second), rcon.SetDeadline(5*time.Second))
		if err1 != nil {
			err = err1 // I'm actually not sure can I just use err here (i'm about var shadowing fyi)
			log.Warn().Err(err1).Str("address", address).Int("attempt", i+1).Msg("retrying connection...")
		} else {
			s.conn = conn
			return nil
		}
		if i < attempts-1 {
			time.Sleep(3 * time.Second)
		}
	}
	log.Err(err).Str("address", address).Int("attempts", attempts).Msg("failed to connect to source rcon")
	return fmt.Errorf("failed to connect: %w", err)
}

func (s *Source) Disconnect() {
	if s.conn == nil {
		return
	}
	s.conn.Close()
	s.conn = nil
}

func (s *Source) Execute(cmd string) (string, error) {
	if s.conn == nil {
		return "", iRcon.ErrNotConnected
	}

	response, err := s.conn.Execute(cmd)
	if err == nil {
		return response, nil
	}

	if errors.Is(err, rcon.ErrCommandEmpty) ||
		errors.Is(err, rcon.ErrCommandTooLong) {
		return "", fmt.Errorf("invalid command: %w", err)
	}

	log.Warn().Err(err).Msg("rcon execution failed, reconnecting...")

	s.Disconnect()
	if reconnectErr := s.Connect(); reconnectErr != nil {
		return "", errors.Join(
			fmt.Errorf("failed executing command: %w", err),
			fmt.Errorf("failed reconnecting: %w", reconnectErr),
		)
	}

	response, err = s.conn.Execute(cmd)
	if err != nil {
		s.Disconnect()
		return "", fmt.Errorf("failed executing command after reconnect: %w", err)
	}

	return response, nil
}
