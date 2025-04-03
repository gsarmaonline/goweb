package server

import (
	"context"

	"github.com/gin-gonic/gin"
)

type (
	Server struct {
		ctx context.Context
		cfg *ServerConfig

		apiEngine *gin.Engine
	}

	ServerConfig struct {
		Host string `json:"host"`
		Port string `json:"port"`
	}
)

func NewServer(ctx context.Context, cfg *ServerConfig) (srv *Server, err error) {
	srv = &Server{
		ctx: ctx,
		cfg: cfg,
	}
	srv.apiEngine = gin.Default()
	return
}

func (srv *Server) Run() (err error) {
	if err = srv.apiEngine.Run(); err != nil {
		return
	}
	return
}
