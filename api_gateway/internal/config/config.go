package config

import (
	"errors"
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env     string         `yaml:"env" env-default:"local"`
	GRPC    *GRPCConfig    `yaml:"grpc"`
	HTTP    *HTTPConfig    `yaml:"http"`
	Auth    *AuthConfig    `yaml:"auth"`
	Kafka   *KafkaConfig   `yaml:"kafka"`
	Storage *StorageConfig `yaml:"storage"`
}

type GRPCConfig struct {
	Timeout time.Duration `yaml:"timeout"`
}

type HTTPConfig struct {
	Port            int           `yaml:"port" env-default:"8080"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env-default:"10s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env-default:"10s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env-default:"5s"`
}

type AuthConfig struct {
	ServiceAddress string        `yaml:"service_address"`
	Timeout        time.Duration `yaml:"timeout"`
}

type KafkaConfig struct {
	Brokers         []string `yaml:"brokers" env-required:"true"`
	TextTopic       string   `yaml:"text_topic" env-default:"content.text"`
	ImageTopic      string   `yaml:"image_topic" env-default:"content.images"`
	DeadLetterTopic string   `yaml:"dead_letter_topic" env-default:"content.dead_letter"`
}

type StorageConfig struct {
	ServiceAddress string        `yaml:"service_address"`
	Timeout        time.Duration `yaml:"timeout"`
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
