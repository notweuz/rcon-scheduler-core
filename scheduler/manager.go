package scheduler

import (
	"fmt"
	"log/slog"

	"github.com/notweuz/rcon-scheduler-core/rcon"
	"github.com/notweuz/rcon-scheduler-core/rcon/source"
	"github.com/notweuz/rcon-scheduler-core/server"
)

type Manager struct {
	Clients map[server.ServerKey]rcon.Client
	logger  *slog.Logger
}

func New(logger *slog.Logger) *Manager {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &Manager{
		Clients: make(map[server.ServerKey]rcon.Client),
		logger:  logger,
	}
}

func (m *Manager) Register(srv server.Server) {
	address := fmt.Sprintf("%v:%d", srv.Address, srv.Port)
	m.logger.Info("register new server", "type", srv.Type, "address", address)

	key := server.ServerKey{
		Address: srv.Address,
		Port:    srv.Port,
	}

	clientLogger := m.logger.With("server", address)
	switch srv.Type {
	case server.TypeSource:
		m.Clients[key] = source.New(address, srv.Password, source.WithLogger(clientLogger))
	default:
		m.logger.Warn("unexpected type, fallback to source", "type", srv.Type)
		m.Clients[key] = source.New(address, srv.Password, source.WithLogger(clientLogger))
	}
}

func (m *Manager) Unregister(srv server.Server) {
	address := fmt.Sprintf("%v:%d", srv.Address, srv.Port)
	m.logger.Info("unregistering server", "type", srv.Type, "address", address)

	key := server.ServerKey{
		Address: srv.Address,
		Port:    srv.Port,
	}

	delete(m.Clients, key)
}
