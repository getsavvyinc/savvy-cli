package config

import (
	"encoding/json"
	"os"
)

var (
	DefaultConfigPath = os.ExpandEnv("$HOME/.config/savvy/config.json")
)

type Config struct {
	Token string
}

func (c *Config) Save() error {
	f, err := os.Create(DefaultConfigPath)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(c); err != nil {
		return err
	}
	return nil
}

func LoadFromFile() (*Config, error) {
	f, err := os.Open(DefaultConfigPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var c Config
	if err := json.NewDecoder(f).Decode(&c); err != nil {
		return nil, err
	}
	return &c, nil
}
