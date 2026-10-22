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

var (
	ErrUnsupportedType = errors.New("unsupported server type")
	ErrServerExists    = errors.New("server already exists")
	ErrServerNotFound  = errors.New("server not found")
)

type Scheduler struct {
	clients map[server.Key]rcon.Client
	cron    gocron.Scheduler
	logger  zerolog.Logger

	mu sync.RWMutex
}

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

func (s *Scheduler) Execute(ctx context.Context, key server.Key, cmd string) (string, error) {
	s.mu.RLock()
	client, ok := s.clients[key]
	s.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("execute on %s: %w", key.String(), ErrServerNotFound)
	}
	return client.Execute(ctx, cmd)
}

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
	client.Disconnect()

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

func (s *Scheduler) Start() {
	s.cron.Start()
}

func (s *Scheduler) Shutdown() error {
	err := s.cron.Shutdown()

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, c := range s.clients {
		c.Disconnect()
	}
	return err
}
