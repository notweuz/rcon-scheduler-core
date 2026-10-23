// Package source implements an RCON client for the Websocket RCON protocol.
package websocket

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	gorcon "github.com/gorcon/websocket"
	"github.com/notweuz/rcon-scheduler-core/rcon"
	"github.com/rs/zerolog"
)

// Option configures a [Websocket] created with [New].
type Option func(*Websocket)

// WithLogger sets the logger for connection and reconnection events.
// By default all log output is discarded.
func WithLogger(l zerolog.Logger) Option {
	return func(w *Websocket) {
		w.logger = l
	}
}

// Websocket is a Websocket RCON client. It implements [rcon.Client] and is safe for
// concurrent use: commands are sent one at a time over a single connection.
type Websocket struct {
	address  string
	password string
	logger   zerolog.Logger

	mu     sync.Mutex // guards the fields below
	conn   *gorcon.Conn
	closed bool
}

const attempts = 5

// New returns a client for the server at address ("host:port").
// No connection is made until [Websocket.Connect] or [Websocket.Execute] is called.
func New(address, password string, opts ...Option) *Websocket {
	w := &Websocket{
		address:  address,
		password: password,
		logger:   zerolog.Nop(),
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

// Connect opens the connection if it is not open yet. It makes up to 5 attempts,
// 5 seconds apart; the pause between attempts is interrupted when ctx is cancelled.
//
// Calling Connect is optional: [Websocket.Execute] connects on demand.
// After [Websocket.Close] it returns [rcon.ErrClosed].
func (w *Websocket) Connect(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.connect(ctx)
}

func (w *Websocket) connect(ctx context.Context) error {
	if w.closed {
		return rcon.ErrClosed
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	if w.conn != nil {
		return nil
	}

	var err error
	for i := range attempts {
		var conn *gorcon.Conn
		conn, err = gorcon.Dial(w.address, w.password, gorcon.SetDialTimeout(5*time.Second), gorcon.SetDeadline(5*time.Second))
		if err == nil {
			w.conn = conn
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
	return fmt.Errorf("connect %s: %w", w.address, err)
}

// Disconnect closes the current connection. A later [Websocket.Execute] reconnects.
// Use [Websocket.Close] to close the client permanently.
func (w *Websocket) Disconnect() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.disconnect()
}

func (w *Websocket) disconnect() {
	if w.conn == nil {
		return
	}
	w.conn.Close()
	w.conn = nil
}

// Execute sends cmd to the server and returns its response, connecting first
// if needed (see [Websocket.Connect]). Concurrent calls wait for each other.
//
// If the command fails because of a network error, Execute reconnects and
// retries it once. Empty or too long commands fail without a retry.
// After [Websocket.Close] it returns [rcon.ErrClosed].
func (w *Websocket) Execute(ctx context.Context, cmd string) (string, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.connect(ctx); err != nil {
		return "", fmt.Errorf("failed to connect on execute: %w", err)
	}

	response, err := w.conn.Execute(cmd)
	if err == nil {
		return response, nil
	}

	if errors.Is(err, gorcon.ErrCommandEmpty) ||
		errors.Is(err, gorcon.ErrCommandTooLong) {
		return "", fmt.Errorf("invalid command: %w", err)
	}

	w.logger.Warn().Err(err).Str("address", w.address).Msg("rcon execution failed, reconnecting")

	w.disconnect()
	if reconnectErr := w.connect(ctx); reconnectErr != nil {
		return "", errors.Join(
			fmt.Errorf("failed executing command: %w", err),
			fmt.Errorf("failed reconnecting: %w", reconnectErr),
		)
	}

	response, err = w.conn.Execute(cmd)
	if err != nil {
		w.disconnect()
		return "", fmt.Errorf("failed executing command after reconnect: %w", err)
	}

	return response, nil
}

// Close closes the connection permanently. After Close, [Websocket.Connect] and
// [Websocket.Execute] return [rcon.ErrClosed]. It is safe to call Close more than once.
func (w *Websocket) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.closed = true
	w.disconnect()
	return nil
}
