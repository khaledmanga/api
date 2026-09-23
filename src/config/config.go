package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func LoadConfig() (*Config, error) {
	data, err := os.ReadFile("config.yml")

	if err != nil {
		return nil, err
	}

	var cfg Config

	if err != yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
