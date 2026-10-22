package manager

import (
	"fmt"

	"github.com/notweuz/rcon-scheduler-core/rcon"
	"github.com/notweuz/rcon-scheduler-core/rcon/source"
	"github.com/notweuz/rcon-scheduler-core/server"

	"github.com/rs/zerolog/log"
)

type Manager struct {
	Clients map[server.ServerKey]rcon.Client
}

func New() *Manager {
	return &Manager{
		Clients: make(map[server.ServerKey]rcon.Client),
	}
}

func (m *Manager) Register(srv server.Server) {
	log.Info().Str("type", string(srv.Type)).Str("address", fmt.Sprintf("%v:%d", srv.Address, srv.Port)).Msg("register new server")

	key := server.ServerKey{
		Address: srv.Address,
		Port:    srv.Port,
	}

	address := fmt.Sprintf("%v:%d", srv.Address, srv.Port)
	switch srv.Type {
	case server.TypeSource:
		m.Clients[key] = source.New(address, srv.Password)
	default:
		log.Warn().Str("type", string(srv.Type)).Msg("unexpected type, fallback to source")
		m.Clients[key] = source.New(address, srv.Password)
	}
}

func (m *Manager) Unregister(srv server.Server) {
	log.Info().Str("type", string(srv.Type)).Str("address", fmt.Sprintf("%v:%d", srv.Address, srv.Port)).Msg("unregistering server")

	key := server.ServerKey{
		Address: srv.Address,
		Port:    srv.Port,
	}

	delete(m.Clients, key)
}
