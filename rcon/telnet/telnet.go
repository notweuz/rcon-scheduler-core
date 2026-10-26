// Package telnet implements an RCON client for the Telnet RCON protocol.
package telnet

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	gorcon "github.com/gorcon/telnet"
	"github.com/notweuz/rcon-scheduler-core/rcon"
	"github.com/rs/zerolog"
)

// Option configures a [Telnet] created with [New]
type Option func(*Telnet)

// WithLogger sets the logger for connection and reconnection events.
// By default all log output is discarded.
func WithLogger(l zerolog.Logger) Option {
	return func(t *Telnet) {
		t.logger = l
	}
}

// Telnet is a Telnet RCON client. It implements [rcon.Client] and is safe for
// concurrent use: commands are sent one at a time over a single connection.
type Telnet struct {
	address  string
	password string
	logger   zerolog.Logger

	mu     sync.Mutex
	conn   *gorcon.Conn
	closed bool
}

const attempts = 5

// New returns a client for the server at address ("host:port").
// No connection is made until [Telnet.Connect] or [Telnet.Execute] is called.
func New(address, password string, opts ...Option) *Telnet {
	t := &Telnet{
		address:  address,
		password: password,
		logger:   zerolog.Nop(),
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// Connect opens the connection if it is not open yet. It makes up to 5 attempts,
// 5 seconds apart; the pause between attempts is interrupted when ctx is cancelled.
//
// Calling Connect is optional: [Telnet.Execute] connects on demand.
// After [Telnet.Close] it returns [rcon.ErrClosed].
func (t *Telnet) Connect(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.connect(ctx)
}

func (t *Telnet) connect(ctx context.Context) error {
	if t.closed {
		return rcon.ErrClosed
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	if t.conn != nil {
		return nil
	}

	var err error
	for i := range attempts {
		var conn *gorcon.Conn
		conn, err = gorcon.Dial(t.address, t.password, gorcon.SetDialTimeout(5*time.Second))
		if err == nil {
			t.conn = conn
			return nil
		}
		if i < attempts-1 {
			select {
			case <-time.After(5*time.Second):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return fmt.Errorf("connect %s: %w", t.address, err)
}

// Disconnect closes the current connection. A later [Telnet.Execute] reconnects.
// Use [Telnet.Close] to close the client permanently.
func (t *Telnet) Disconnect() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.disconnect()
}

func (t *Telnet) disconnect() {
	if t.conn == nil {
		return
	}
	t.conn.Close()
	t.conn = nil
}

// Execute sends cmd to the server and returns its response, connecting first
// if needed (see [Telnet.Connect]). Concurrent calls wait for each other.
//
// If the command fails because of a network error, Execute reconnects and
// retries it once. Empty or too long commands fail without a retry.
// After [Telnet.Close] it returns [rcon.ErrClosed].
func (t *Telnet) Execute(ctx context.Context, cmd string) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if err := t.connect(ctx); err != nil {
		return "", fmt.Errorf("failed to connect on execute: %w", err)
	}

	response, err := t.conn.Execute(cmd)
	if err == nil {
		return response, nil
	}

	if errors.Is(err, gorcon.ErrCommandEmpty) ||
		errors.Is(err, gorcon.ErrCommandTooLong) {
		return "", fmt.Errorf("invalid command: %w", err)
	}

	t.logger.Warn().Err(err).Str("address", t.address).Msg("rcon execution failed, reconnecting")
	
	t.disconnect()
	if reconnectErr := t.connect(ctx); reconnectErr != nil {
		return "", errors.Join(
			fmt.Errorf("failed executing command: %w", err),
			fmt.Errorf("failed reconnecting: %w", reconnectErr),
		)
	}

	response, err = t.conn.Execute(cmd)
	if err != nil {
		t.disconnect()
		return "", fmt.Errorf("failed executing command after reconnect: %w", err)
	}

	return response, nil
}

// Close closes the connection permanently. After Close, [Telnet.Connect] and
// [Telnet.Execute] return [rcon.ErrClosed]. It is safe to call Close more than once.
func (t *Telnet) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.closed = true
	t.disconnect()
}