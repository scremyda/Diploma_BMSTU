package main

import (
	"diploma/alerter/repo"
	"diploma/alerter/telegram"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Database DatabaseConfig          `yaml:"database"`
	Telegram telegram.TelegramConfig `yaml:"telegram"`
	Queue    repo.QueueConfig        `yaml:"queue"`
}

type DatabaseConfig struct {
	DBUser string `yaml:"db_user"`
	DBPass string `yaml:"db_pass"`
	DBHost string `yaml:"db_host"`
	DBPort int    `yaml:"db_port"`
	DBName string `yaml:"db_name"`
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
