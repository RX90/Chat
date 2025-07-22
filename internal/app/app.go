package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RX90/Chat/config"
	"github.com/RX90/Chat/internal/db/postgres"
	"github.com/RX90/Chat/internal/handler"
	"github.com/RX90/Chat/internal/repo"
	"github.com/RX90/Chat/internal/server"
	"github.com/RX90/Chat/internal/service"
)

type App struct {
	cfg    *config.Config
	server *server.Server
}

func NewApp() (*App, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	db, err := postgres.NewPostgresDB(cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	repo := repo.NewRepo(db)
	service := service.NewService(repo)
	handler := handler.NewHandler(service)

	r, err := server.NewRouter(handler)
	if err != nil {
		return nil, fmt.Errorf("failed to get router: %w", err)
	}
	srv := server.NewServer(cfg.Server, r)

	return &App{
		cfg:    cfg,
		server: srv,
	}, nil
}

func (a *App) Run() {
	go func() {
		if err := a.server.Run(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server encountered an unexpected error: %v", err)
		}
	}()

	log.Printf("Chat started on :%s\n", a.cfg.Server.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Shutdown server: %v", err)
	}
}
