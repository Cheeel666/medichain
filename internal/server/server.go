package server

import (
	"8sem/diploma/medichain/internal/repo"
	"8sem/diploma/medichain/internal/service"
)

type Server struct {
	Services     service.Services
	Repositories repo.Repositories
	//	TODO: conn
	//	TODO: log
}

func NewServer() *Server {
	repos :=
}
