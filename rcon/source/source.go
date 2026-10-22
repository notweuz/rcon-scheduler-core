package source

import (
	"context"
	"errors"
	"fmt"
	"time"

	iRcon "github.com/notweuz/rcon-scheduler-core/rcon"

	"github.com/gorcon/rcon"
	"github.com/rs/zerolog/log"
)

// TODO: apparently this needs mutex so it won't read other jobs while it's reading the current one
type Source struct {
	address  string
	password string

	conn *rcon.Conn
}

const attempts = 5

func New(address, password string) *Source {
	return &Source{
		address:  address,
		password: password,
	}
}

func (s *Source) Connect(ctx context.Context) error {
	if s.conn != nil {
		return nil
	}

	var err error
	for i := range attempts {
		var conn *rcon.Conn
		conn, err = rcon.Dial(s.address, s.password, rcon.SetDialTimeout(5*time.Second), rcon.SetDeadline(5*time.Second))
		if err == nil {
			s.conn = conn
			return nil
		}
		if i < attempts-1 {
			select {
			case <-time.After(5 * time.Second):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return fmt.Errorf("connect %s: %w", s.address, err)
}

func (s *Source) Disconnect() {
	if s.conn == nil {
		return
	}
	s.conn.Close()
	s.conn = nil
}

func (s *Source) Execute(ctx context.Context, cmd string) (string, error) {
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
	if reconnectErr := s.Connect(ctx); reconnectErr != nil {
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
