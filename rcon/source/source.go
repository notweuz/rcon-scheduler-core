// Package source implements an RCON client for the Source RCON protocol,
// which is also used by Minecraft.
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

// Option configures a [Source] created with [New].
type Option func(*Source)

// WithLogger sets the logger for connection and reconnection events.
// By default all log output is discarded.
func WithLogger(l zerolog.Logger) Option {
	return func(s *Source) {
		s.logger = l
	}
}

// Source is a Source RCON client. It implements [rcon.Client] and is safe for
// concurrent use: commands are sent one at a time over a single connection.
type Source struct {
	address  string
	password string
	logger   zerolog.Logger

	mu     sync.Mutex // guards the fields below
	conn   *gorcon.Conn
	closed bool
}

const attempts = 5

// New returns a client for the server at address ("host:port").
// No connection is made until [Source.Connect] or [Source.Execute] is called.
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

// Connect opens the connection if it is not open yet. It makes up to 5 attempts,
// 5 seconds apart; the pause between attempts is interrupted when ctx is cancelled.
//
// Calling Connect is optional: [Source.Execute] connects on demand.
// After [Source.Close] it returns [rcon.ErrClosed].
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

// Disconnect closes the current connection. A later [Source.Execute] reconnects.
// Use [Source.Close] to close the client permanently.
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

// Execute sends cmd to the server and returns its response, connecting first
// if needed (see [Source.Connect]). Concurrent calls wait for each other.
//
// If the command fails because of a network error, Execute reconnects and
// retries it once. Empty or too long commands fail without a retry.
// After [Source.Close] it returns [rcon.ErrClosed].
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

// Close closes the connection permanently. After Close, [Source.Connect] and
// [Source.Execute] return [rcon.ErrClosed]. It is safe to call Close more than once.
func (s *Source) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.closed = true
	s.disconnect()
	return nil
}
