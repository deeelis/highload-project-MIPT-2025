package config

import (
	"errors"
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

type Config struct {
	Env   string       `yaml:"env" env-default:"local"`
	Kafka *KafkaConfig `yaml:"kafka"`
}

type KafkaConfig struct {
	Brokers     []string `yaml:"brokers" env-required:"true"`
	InputTopic  string   `yaml:"input_topic" env-default:"content.text"`
	ResultTopic string   `yaml:"result_topic" env-default:"content.text.results"`
	ErrorTopic  string   `yaml:"error_topic" env-default:"content.dead_letter"`
	GroupID     string   `yaml:"group_id" env-default:"text_analyzer_group"`
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
