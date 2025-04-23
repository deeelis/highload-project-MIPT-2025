package config

import (
	"errors"
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env      string         `yaml:"env" env-default:"local"`
	GRPC     GRPCConfig     `yaml:"grpc"`
	Database DatabaseConfig `yaml:"database"`
	Token    TokenConfig    `yaml:"token"`
}

type TokenConfig struct {
	TokenTTL        time.Duration `yaml:"token_ttl" env-default:"1h"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl" env-default:"168h"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type DatabaseConfig struct {
	DSN string `yaml:"dsn"`
}

func MustLoad() (*Config, error) {
	configPath := fetchConfigPath()
	if configPath == "" {
		return nil, errors.New("config path is empty")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, errors.New("config path isn't exist: " + err.Error())
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, errors.New("can't read config: " + err.Error())
	}

	return &cfg, nil
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
