package source

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	gorcon "github.com/gorcon/rcon"
	"github.com/notweuz/rcon-scheduler-core/rcon"
	"github.com/rs/zerolog"
)

type Option func(*Source)

func WithLogger(l zerolog.Logger) Option {
	return func(s *Source) {
		s.logger = l
	}
}

type Source struct {
	address  string
	password string
	closed   bool

	logger zerolog.Logger
	mu     sync.Mutex
	conn   *gorcon.Conn
}

const attempts = 5

func New(address, password string, opts ...Option) *Source {
	s := &Source{
		address:  address,
		password: password,
		logger:   zerolog.Nop(),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Source) Connect(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.connect(ctx)
}

func (s *Source) connect(ctx context.Context) error {
	if s.closed {
		return rcon.ErrClosed
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	if s.conn != nil {
		return nil
	}

	var err error
	for i := range attempts {
		var conn *gorcon.Conn
		conn, err = gorcon.Dial(s.address, s.password, gorcon.SetDialTimeout(5*time.Second), gorcon.SetDeadline(5*time.Second))
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
	s.mu.Lock()
	defer s.mu.Unlock()

	s.disconnect()
}

func (s *Source) disconnect() {
	if s.conn == nil {
		return
	}
	s.conn.Close()
	s.conn = nil
}

func (s *Source) Execute(ctx context.Context, cmd string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.connect(ctx); err != nil {
		return "", fmt.Errorf("failed to connect on execute: %w", err)
	}

	response, err := s.conn.Execute(cmd)
	if err == nil {
		return response, nil
	}

	if errors.Is(err, gorcon.ErrCommandEmpty) ||
		errors.Is(err, gorcon.ErrCommandTooLong) {
		return "", fmt.Errorf("invalid command: %w", err)
	}

	s.logger.Warn().Err(err).Str("address", s.address).Msg("rcon execution failed, reconnecting")

	s.disconnect()
	if reconnectErr := s.connect(ctx); reconnectErr != nil {
		return "", errors.Join(
			fmt.Errorf("failed executing command: %w", err),
			fmt.Errorf("failed reconnecting: %w", reconnectErr),
		)
	}

	response, err = s.conn.Execute(cmd)
	if err != nil {
		s.disconnect()
		return "", fmt.Errorf("failed executing command after reconnect: %w", err)
	}

	return response, nil
}

func (s *Source) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.closed = true
	s.disconnect()
	return nil
}