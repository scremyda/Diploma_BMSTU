package main

import (
	"diploma/alerter/db"
	"diploma/alerter/repo"
	"diploma/alerter/telegram"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database db.Config       `yaml:"database"`
	Telegram telegram.Config `yaml:"telegram"`
	Queue    repo.Config     `yaml:"queue"`
}

func ReadConf(cfgPath string) (*Config, error) {
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
