package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/notweuz/rcon-scheduler-core/rcon"
	"github.com/notweuz/rcon-scheduler-core/rcon/source"
	"github.com/notweuz/rcon-scheduler-core/server"
	"github.com/rs/zerolog"
)

// Scheduler runs [job.Job]s on registered servers according to their cron schedules.
// Create it with [New].
type Scheduler struct {
	clients map[server.Key]rcon.Client
	cron    gocron.Scheduler
	logger  zerolog.Logger

	mu sync.RWMutex
}

// New creates a Scheduler. Jobs don't run until [Scheduler.Start] is called.
//
// By default the Scheduler uses UTC and discards all log output;
// see [WithLocation] and [WithLogger].
func New(opts ...Option) (*Scheduler, error) {
	o := options{
		location: time.UTC,
		logger:   zerolog.Nop(),
	}
	for _, opt := range opts {
		opt(&o)
	}

	c, err := gocron.NewScheduler(gocron.WithLocation(o.location))
	if err != nil {
		return nil, fmt.Errorf("create cron scheduler: %w", err)
	}
	return &Scheduler{
		cron:    c,
		logger:  o.logger,
		clients: make(map[server.Key]rcon.Client),
	}, nil
}

// Register adds srv to the scheduler and creates an RCON client for it.
// No connection is made until the first command is executed.
//
// It returns an error wrapping:
//   - the error from [server.Server.Validate] if srv is invalid;
//   - [ErrUnsupportedType] if no client implementation exists for srv.Type;
//   - [ErrServerExists] if a server with the same key is already registered.
func (s *Scheduler) Register(srv server.Server) error {
	if err := srv.Validate(); err != nil {
		return fmt.Errorf("validation failed during addition of server %v: %w", srv.Key(), err)
	}

	client, err := s.newClient(srv)
	if err != nil {
		return fmt.Errorf("add server %s: %w", srv.Key(), err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.clients[srv.Key()]; ok {
		return fmt.Errorf("add server %s: %w", srv.Key(), ErrServerExists)
	}
	s.clients[srv.Key()] = client

	s.logger.Info().Str("server", srv.Key().String()).Str("type", string(srv.Type)).Msg("server added")
	return nil
}

// Execute runs cmd on the server identified by key and returns the server's response.
//
// The command uses the same connection as scheduled jobs and waits for any
// command currently in progress on that server. If the connection is down,
// Execute reconnects before sending the command.
//
// It returns [ErrServerNotFound] if the server is not registered.
func (s *Scheduler) Execute(ctx context.Context, key server.Key, cmd string) (string, error) {
	s.mu.RLock()
	client, ok := s.clients[key]
	s.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("execute on %s: %w", key.String(), ErrServerNotFound)
	}
	resp, err := client.Execute(ctx, cmd)
	if errors.Is(err, rcon.ErrClosed) {
		return "", fmt.Errorf("execute on %s: %w", key, ErrServerNotFound)
	}
	return resp, err
}

// Unregister removes the server identified by key, removes all of its jobs
// and closes its connection. Jobs that are running at that moment have their
// context cancelled.
//
// It returns [ErrServerNotFound] if no such server is registered.
func (s *Scheduler) Unregister(key server.Key) error {
	s.mu.Lock()
	client, ok := s.clients[key]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("remove server %s: %w", key.String(), ErrServerNotFound)
	}
	delete(s.clients, key)
	s.mu.Unlock()

	s.cron.RemoveByTags(key.String())
	client.Close()

	s.logger.Info().Str("server", key.String()).Msg("server removed")
	return nil
}

func (s *Scheduler) newClient(srv server.Server) (rcon.Client, error) {
	clientLogger := s.logger.With().Str("server", srv.Key().String()).Logger()
	switch srv.Type {
	case server.TypeSource:
		return source.New(srv.Key().String(), srv.Password, source.WithLogger(clientLogger)), nil
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedType, srv.Type)
	}
}

// Start starts running scheduled jobs. Jobs added after Start are scheduled immediately.
func (s *Scheduler) Start() {
	s.cron.Start()
}

// Shutdown stops the schedule, cancels the context of running jobs, waits for
// them to finish and closes all connections. The Scheduler cannot be started again.
func (s *Scheduler) Shutdown() error {
	err := s.cron.Shutdown()

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, c := range s.clients {
		c.Close()
	}
	return err
}
