package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const DefaultConfigFileName = "config.json"

var (
	DefaultConfigDir      = os.ExpandEnv("$HOME/.config/savvy")
	DefaultConfigFilePath = filepath.Join(DefaultConfigDir, DefaultConfigFileName)
)

type Config struct {
	Token         string `json:"token"`
	OpenAIBaseURL string `json:"openai_base_url"`
}

func (c *Config) Save() error {
	if _, err := os.Stat(DefaultConfigDir); os.IsNotExist(err) {
		if err := os.MkdirAll(DefaultConfigDir, 0755); err != nil {
			return err
		}
	}

	f, err := os.Create(DefaultConfigFilePath)
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
	f, err := os.Open(DefaultConfigFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var c Config
	if err := json.NewDecoder(f).Decode(&c); err != nil {
		return nil, err
	}

	if err := c.validate(); err != nil {
		return nil, err
	}

	return &c, nil
}

var ErrMissingToken = errors.New("missing or empty token")

func (c *Config) validate() error {
	if c.Token == "" {
		return ErrMissingToken
	}
	return nil
}
