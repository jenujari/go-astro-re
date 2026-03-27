package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Engine   EngineConfig   `mapstructure:"engine"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

type ServerConfig struct {
	HTTPAddr          string        `mapstructure:"http_addr"`
	ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout"`
	ShutdownTimeout   time.Duration `mapstructure:"shutdown_timeout"`
}

type DatabaseConfig struct {
	Enabled         bool          `mapstructure:"enabled"`
	URL             string        `mapstructure:"url"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type EngineConfig struct {
	WorkerCount    int           `mapstructure:"worker_count"`
	RequestTimeout time.Duration `mapstructure:"request_timeout"`
	PersistResults bool          `mapstructure:"persist_results"`
}

type LoggingConfig struct {
	Level string `mapstructure:"level"`
}

func Load(path string) (Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config %s: %w", path, err)
	}

	applyDefaults(&cfg)
	return cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Server.HTTPAddr == "" {
		cfg.Server.HTTPAddr = ":8080"
	}
	if cfg.Server.ReadHeaderTimeout <= 0 {
		cfg.Server.ReadHeaderTimeout = 5 * time.Second
	}
	if cfg.Server.ShutdownTimeout <= 0 {
		cfg.Server.ShutdownTimeout = 10 * time.Second
	}
	if cfg.Database.MaxOpenConns <= 0 {
		cfg.Database.MaxOpenConns = 10
	}
	if cfg.Database.MaxIdleConns <= 0 {
		cfg.Database.MaxIdleConns = 5
	}
	if cfg.Database.ConnMaxLifetime <= 0 {
		cfg.Database.ConnMaxLifetime = 15 * time.Minute
	}
	if cfg.Engine.WorkerCount <= 0 {
		cfg.Engine.WorkerCount = 4
	}
	if cfg.Engine.RequestTimeout <= 0 {
		cfg.Engine.RequestTimeout = 10 * time.Second
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "INFO"
	}
}
