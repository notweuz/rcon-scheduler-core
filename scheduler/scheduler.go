package scheduler

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/notweuz/rcon-scheduler-core/rcon"
	"github.com/notweuz/rcon-scheduler-core/rcon/source"
	"github.com/notweuz/rcon-scheduler-core/server"
)

var (
	ErrUnsupportedType = errors.New("unsupported server type")
	ErrServerExists    = errors.New("same server already exist")
)

type Scheduler struct {
	clients map[server.Key]rcon.Client
	logger  *slog.Logger

	mu sync.RWMutex
}

func (s *Scheduler) Register(srv server.Server) error {
	if err := srv.Validate(); err != nil {
		return fmt.Errorf("validation failed during addition of server %v: %w", srv.Address, err)
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

	s.logger.Info("server added", "server", srv.Key().String(), "type", srv.Type)
	return nil
}

func (s *Scheduler) newClient(srv server.Server) (rcon.Client, error) {
	clientLogger := s.logger.With("server", srv.Key().String())
	switch srv.Type {
	case server.TypeSource:
		return source.New(srv.Key().String(), srv.Password, source.WithLogger(clientLogger)), nil
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedType, srv.Type)
	}
}
