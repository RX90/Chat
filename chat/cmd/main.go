package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RX90/Chat/chat"
	"github.com/RX90/Chat/db"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

func main() {
	if err := initConfig(); err != nil {
		log.Fatalf("can't init config: %v", err)
	}

	if err := godotenv.Load(); err != nil {
		log.Fatalf("can't load .env: %v", err)
	}

	server := chat.NewServer()

	postgres, err := db.NewPostgresDB(db.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
		Password: os.Getenv("DB_PASSWORD"),
	})
	if err != nil {
		log.Fatalf("can't start db: %v", err)
	}

	hub := chat.NewHub()
	go hub.Run()
	
	c := &chat.Chat{
		Server: server,
		Hub:    hub,
		DB:     postgres,
	}

	go func() {
		c.Server.Run(viper.GetString("chat.port"), c.InitRoutes())
	}()

	log.Printf("chat module started on :%s\n", viper.GetString("chat.port"))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Println("shutting down chat module")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Server.Shutdown(ctx); err != nil {
		log.Fatalf("error occurred during chat server shutdown: %v", err)
	}

	if err := c.DB.Close(); err != nil {
		log.Fatalf("error occurred while closing db connection (chat): %v", err)
	}
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	required := []string{
		"db.username",
		"db.host",
		"db.port",
		"db.dbname",
		"db.sslmode",
	}
	missing := []string{}

	for _, key := range required {
		if viper.GetString(key) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required config values: %v", missing)
	}

	return nil
}
