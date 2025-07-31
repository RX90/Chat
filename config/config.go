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
	cfg.DB.Password = os.Getenv("POSTGRES_PASSWORD")

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return &cfg, nil
}

func (cfg *Config) Validate() error {
	if cfg.Server == nil {
		return errors.New("server config is required")
	}
	if cfg.Server.Port == "" {
		return errors.New("server port is required")
	}
	if cfg.Server.MaxHeaderBytes <= 0 {
		return errors.New("maxHeaderBytes must be positive")
	}
	if cfg.Server.ReadTimeout <= 0 {
		return errors.New("readTimeout must be positive")
	}
	if cfg.Server.WriteTimeout <= 0 {
		return errors.New("writeTimeout must be positive")
	}

	if cfg.DB == nil {
		return errors.New("db config is required")
	}
	if cfg.DB.Host == "" {
		return errors.New("db host is required")
	}
	if cfg.DB.Port == "" {
		return errors.New("db port is required")
	}
	if cfg.DB.Username == "" {
		return errors.New("db username is required")
	}
	if cfg.DB.Password == "" {
		return errors.New("db password is required")
	}
	if cfg.DB.DBName == "" {
		return errors.New("db name is required")
	}
	if cfg.DB.SSLMode == "" {
		return errors.New("db sslmode is required")
	}

	return nil
}
