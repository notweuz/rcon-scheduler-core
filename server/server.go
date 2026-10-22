package server

type ServerKey struct {
	Address string
	Port    int
}

type Server struct {
	Address  string
	Port     int
	Password string
	Type     Type
}

func New(address string, port int, password string, serverType Type) *Server {
	return &Server{
		Address:  address,
		Port:     port,
		Password: password,
		Type:     serverType,
	}
}
