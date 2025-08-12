package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

var LogLevel uint8

type Config struct {
	App      AppSettings  `yaml:"app"`
	Logger   LoggerConfig `yaml:"logger"`
	Server   ServerConfig `yaml:"server"`
	DataBase DBConfig     `yaml:"database"`
	Redis    RedisConfig  `yaml:"cache"`
	Timeouts Timeouts     `yaml:"timeouts"`
}

type AppSettings struct {
	Name string `yaml:"name"`
	Env  string `yaml:"env" default:"dev"`
}

type LoggerConfig struct {
	Level      string `yaml:"level" default:"debug"`
	Encoding   string `yaml:"encoding" default:"console"`
	OutputPath string `yaml:"outputPath" default:""`
}

type ServerConfig struct {
	Host         string `yaml:"host" default:"0.0.0.0"`
	Port         int    `yaml:"port" default:"8080"`
	ReadTimeout  string `yaml:"readTimeout" default:"5s"`
	WriteTimeout string `yaml:"writeTimeout" default:"10s"`
	IdleTimeout  string `yaml:"idleTimeout" default:"120s"`
}

type DBConfig struct {
	DSN                string `yaml:"dsn"`
	MaxOpenConnections int    `yaml:"25"`
	MaxIdleConnections int    `yaml:"5"`
}

type RedisConfig struct {
	Address      string `yaml:"address" default:"localhost:6379"`
	Password     string `yaml:"password"`
	DB           int    `yaml:"db"`
	DialTimeout  string `yaml:"dialTimeout" default:"500ms"`
	ReadTimeout  string `yaml:"readTimeout" default:"500ms"`
	WriteTimeout string `yaml:"writeTimeout" default:"500ms"`
	PoolSize     int    `yaml:"poolSize" default:"10"`
}

type Timeouts struct {
	ShutdwonGracePeriod   string `yaml:"shutdownGracePeriod" default:"15s"`
	RequestContentTimeout string `yaml:"requestContentTimeout" default:"30s"`
	ExternalAPITimeout    string `yaml:"externalAPITimeout" default:"10s"`
}

func MustLoadConfig(path string) (*Config, error) {
	if path == "" {
		return nil, fmt.Errorf("config path is empty")
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, fmt.Errorf("cannot read config from %s: %w", path, err)
	}

	return &cfg, nil
}
