package main

import (
	"diploma/alerter/consumer"
	"diploma/alerter/db"
	"diploma/alerter/telegram"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database db.Config       `yaml:"database"`
	Telegram telegram.Config `yaml:"telegram"`
	Consumer consumer.Config `yaml:"consumer"`
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
