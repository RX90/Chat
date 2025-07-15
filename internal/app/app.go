package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/RX90/Chat/config"
	"github.com/RX90/Chat/internal/db/postgres"
	"github.com/RX90/Chat/internal/handler"
	"github.com/RX90/Chat/internal/repo"
	"github.com/RX90/Chat/internal/router"
	"github.com/RX90/Chat/internal/server"
	"github.com/RX90/Chat/internal/ws"
)

type App struct {
	cfg     *config.Config
	hub     *ws.Hub
	server  *server.Server
}

func NewApp() (*App, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("load config error: %w", err)
	}

	db, err := postgres.NewPostgresDB(cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	hub := ws.NewHub()

	repo := repo.NewRepo(db)
	handler := handler.NewHandler(hub, repo)

	r, err := router.NewRouter(handler)
	if err != nil {
		return nil, fmt.Errorf("failed to get router: %w", err)
	}
	srv := server.NewServer(cfg.Server, r)

	return &App{
		cfg:     cfg,
		hub:     hub,
		server:  srv,
	}, nil
}

func (a *App) Run() {
	go func() {
		if err := a.server.Run(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server encountered an unexpected error: %v", err)
		}
	}()

	go a.hub.Run()

	log.Printf("Chat started on :%s\n", a.cfg.Server.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Println("Shutting down HTTP server...")
	if err := a.server.Shutdown(context.Background()); err != nil {
		log.Printf("Error during HTTP server shutdown: %v\n", err)
	}

	log.Println("Shutting down WebSocket hub...")
	a.hub.Shutdown()
}
