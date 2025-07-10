package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	ServerCfg *ServerConfig `yaml:"server"`
}

type ServerConfig struct {
	Port           string        `yaml:"port"`
	MaxHeaderBytes int           `yaml:"maxHeaderBytes"`
	ReadTimeout    time.Duration `yaml:"readTimeout"`
	WriteTimeout   time.Duration `yaml:"writeTimeout"`
}

func NewConfig() *Config {
	return &Config{
		&ServerConfig{
			Port:           viper.GetString("server.port"),
			MaxHeaderBytes: viper.GetInt("server.maxHeaderBytes"),
			ReadTimeout:    viper.GetDuration("server.readTimeout"),
			WriteTimeout:   viper.GetDuration("server.writeTimeout"),
		},
	}
}

func InitConfig() error {
	viper.AddConfigPath("config")
	viper.SetConfigName("config")

	required := []string{
		"server.port",
		"server.maxHeaderBytes",
		"server.readTimeout",
		"server.writeTimeout",
	}
	missing := []string{}

	for _, key := range required {
		if viper.Get(key) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required config values: %v", missing)
	}

	return viper.ReadInConfig()
}