package server

import (
	"context"
	"net/http"
	"time"

	"github.com/RX90/Chat/config"
	"github.com/gin-gonic/gin"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(cfg *config.ServerConfig, router *gin.Engine) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:           ":" + cfg.Port,
			Handler:        router,
			MaxHeaderBytes: cfg.MaxHeaderBytes,
			ReadTimeout:    cfg.ReadTimeout * time.Second,
			WriteTimeout:   cfg.WriteTimeout * time.Second,
		},
	}
}

func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
