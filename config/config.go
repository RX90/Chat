package config

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

//go:embed config.yaml
var EmbeddedConfig []byte

type Config struct {
	Server *ServerConfig `mapstructure:"server"`
	DB     *DBConfig     `mapstructure:"db"`
}

type ServerConfig struct {
	Port           string        `mapstructure:"port"`
	MaxHeaderBytes int           `mapstructure:"maxHeaderBytes"`
	ReadTimeout    time.Duration `mapstructure:"readTimeout"`
	WriteTimeout   time.Duration `mapstructure:"writeTimeout"`
}

type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigType("yaml")

	if err := viper.ReadConfig(bytes.NewReader(EmbeddedConfig)); err != nil {
		return nil, fmt.Errorf("failed to read embedded config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	cfg.DB.Password = os.Getenv("DB_PASSWORD")
	if cfg.DB.Password == "" {
		return nil, errors.New("env variable DB_PASSWORD is not set")
	}

	return &cfg, nil
}
