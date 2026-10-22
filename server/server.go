package server

import "rcon-scheduler-core/scheduler/job"

type ServerKey struct {
	Address string
	Port    int
}

type Server struct {
	Address  string
	Port     int
	Password string
	Type     Type
	Jobs     []job.Job
}

func New(address string, port int, password string, serverType Type) *Server {
	return &Server{
		Address:  address,
		Port:     port,
		Password: password,
		Type:     serverType,
	}
}
