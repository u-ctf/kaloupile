package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Scheme   string `yaml:"scheme"`
	Domain   string `yaml:"domain"`
	Postgres struct {
		LocalHost string `yaml:"localHost"`
		LocalPort int    `yaml:"localPort"`
		Host      string `yaml:"host"`
		Port      int    `yaml:"port"`
		Admin     struct {
			User     string `yaml:"user"`
			Password string `yaml:"password"`
			Database string `yaml:"database"`
		} `yaml:"admin"`
		Users []struct {
			Name      string   `yaml:"name"`
			Password  string   `yaml:"password"`
			Databases []string `yaml:"databases"`
		} `yaml:"users"`
	} `yaml:"postgres"`
}

func Load() (*Config, error) {
	return LoadFromFile("config.yml")
}

func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}

	return &cfg, nil
}
