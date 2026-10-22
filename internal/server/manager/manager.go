package manager

import (
	"fmt"
	"rcon-scheduler-core/internal/rcon"
	"rcon-scheduler-core/internal/rcon/client"
	"rcon-scheduler-core/internal/server"

	"github.com/rs/zerolog/log"
)

type Manager struct {
	Clients map[server.Server]rcon.Client
}

func New() *Manager {
	return &Manager{}
}

func (m *Manager) Register(srv server.Server) {
	log.Info().Str("type", string(srv.Type)).Str("address", fmt.Sprintf("%v:%d", srv.Address, srv.Port)).Msg("register new server")
	switch srv.Type {
	case server.TypeSource:
		m.Clients[srv] = client.NewSource(srv)
	default:
		log.Warn().Str("type", string(srv.Type)).Msg("unexpected type, fallback to source")
		m.Clients[srv] = client.NewSource(srv)
	}
}

func (m *Manager) Unregister(srv server.Server) {
	log.Info().Str("type", string(srv.Type)).Str("address", fmt.Sprintf("%v:%d", srv.Address, srv.Port)).Msg("unregistering server")
	delete(m.Clients, srv)
}