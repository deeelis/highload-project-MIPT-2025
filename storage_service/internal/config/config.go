package config

import (
	"errors"
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env   string      `yaml:"env"`
	GRPC  GRPCConfig  `yaml:"grpc"`
	Repo  RepoConfig  `yaml:"repo"`
	Cache CacheConfig `yaml:"cache"`
	S3    S3Config    `yaml:"s3"`
	Kafka KafkaConfig `yaml:"kafka"`
}

type CacheConfig struct {
	URL string `yaml:"url"`
}
type RepoConfig struct {
	DSN string `yaml:"dsn"`
}

type S3Config struct {
	Endpoint  string `yaml:"endpoint"`
	Region    string `yaml:"region"`
	Bucket    string `yaml:"bucket"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	URL       string `yaml:"url"`
}

type KafkaConfig struct {
	Brokers    []string `yaml:"brokers" env-required:"true"`
	ImageTopic string   `yaml:"image_topic" env-default:"content.image"`
	TextTopic  string   `yaml:"text_topic" env-default:"content.text"`
	ErrorTopic string   `yaml:"error_topic" env-default:"content.dead_letter"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
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
