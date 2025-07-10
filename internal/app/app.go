package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/RX90/Chat/config"
	"github.com/RX90/Chat/internal/router"
	"github.com/RX90/Chat/internal/server"
	"github.com/RX90/Chat/internal/ws"
)

type App struct {
	cfg    *config.Config
	server *server.Server
	hub    *ws.Hub
	router http.Handler
}

func NewApp() *App {
	if err := config.InitConfig(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	cfg := config.NewConfig()
	srv := server.NewServer()
	hub := ws.NewHub()
	r := router.NewRouter(hub)

	return &App{
		cfg: cfg,
		server: srv,
		hub:    hub,
		router: r,
	}
}

func (a *App) Run() {
	go func() {
		if err := a.server.Run(a.cfg.ServerCfg, a.router); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server encountered an unexpected error: %v", err)
		}
	}()

	go a.hub.Run()

	log.Printf("Chat started on :%s\n", a.cfg.ServerCfg.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Println("Initiating graceful shutdown of HTTP server...")
	if err := a.server.Shutdown(context.Background()); err != nil {
		log.Fatalf("Error during HTTP server shutdown: %v", err)
	}

	log.Println("Shutting down WebSocket hub...")
	a.hub.Shutdown()
}
