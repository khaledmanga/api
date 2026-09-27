package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const configFilePath = "src/config/config.yml"

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Password PasswordConfig `yaml:"password"`
	Redis    RedisConfig    `yaml:"redis"`
	CORS     CORSConfig     `yaml:"cors"`
}

type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	TLS      bool   `yaml:"tls"`
	TLSCA    string `yaml:"tls_ca"`
}

type PasswordConfig struct {
	Cost int `yaml:"cost"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	TLS      bool   `yaml:"tls"`
	TLSCA    string `yaml:"tls_ca"`
}

func LoadConfig() (*Config, error) {
	return loadConfig(configFilePath)
}

func loadConfig(path string) (*Config, error) {
	cfg := &Config{}
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("read config file: %w", err)
	}
	if err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse config file: %w", err)
		}
	}

	if err := applyEnvironment(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func applyEnvironment(configuration *Config) error {
	setStringFromEnv("SERVER_HOST", &configuration.Server.Host)
	setStringFromEnv("DB_HOST", &configuration.Database.Host)
	setStringFromEnv("DB_USER", &configuration.Database.User)
	setStringFromEnv("DB_PASSWORD", &configuration.Database.Password)
	setStringFromEnv("DB_NAME", &configuration.Database.Name)
	setStringFromEnv("DB_TLS_CA", &configuration.Database.TLSCA)
	setStringFromEnv("REDIS_HOST", &configuration.Redis.Host)
	setStringFromEnv("REDIS_PASSWORD", &configuration.Redis.Password)
	setStringFromEnv("REDIS_TLS_CA", &configuration.Redis.TLSCA)

	for _, override := range []struct {
		key    string
		target *bool
	}{
		{"DB_TLS", &configuration.Database.TLS},
		{"REDIS_TLS", &configuration.Redis.TLS},
	} {
		if err := setBoolFromEnv(override.key, override.target); err != nil {
			return err
		}
	}

	for _, override := range []struct {
		key    string
		target *int
	}{
		{"SERVER_PORT", &configuration.Server.Port},
		{"DB_PORT", &configuration.Database.Port},
		{"PASSWORD_COST", &configuration.Password.Cost},
		{"REDIS_PORT", &configuration.Redis.Port},
		{"REDIS_DB", &configuration.Redis.DB},
	} {
		if err := setIntFromEnv(override.key, override.target); err != nil {
			return err
		}
	}
	serverPort, serverPortSet := os.LookupEnv("SERVER_PORT")
	if !serverPortSet || serverPort == "" {
		if err := setIntFromEnv("PORT", &configuration.Server.Port); err != nil {
			return err
		}
	}

	if origins, ok := os.LookupEnv("CORS_ALLOWED_ORIGINS"); ok && origins != "" {
		configuration.CORS.AllowedOrigins = splitEnvironmentList(origins)
	}
	return nil
}

func setStringFromEnv(key string, target *string) {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		*target = value
	}
}

func setIntFromEnv(key string, target *int) error {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return nil
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return fmt.Errorf("parse %s: %w", key, err)
	}
	*target = parsed
	return nil
}

func setBoolFromEnv(key string, target *bool) error {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return nil
	}
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return fmt.Errorf("parse %s: %w", key, err)
	}
	*target = parsed
	return nil
}

func splitEnvironmentList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{}
	}
	values := strings.Split(value, ",")
	result := make([]string, 0, len(values))
	for _, item := range values {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
