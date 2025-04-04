package main

import (
	"diploma/certer/certer"
	"diploma/certer/consumer"
	"diploma/certer/db"
	"diploma/certer/producer"
	"diploma/certer/scheduler"
	"diploma/certer/setter"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database  db.Config        `yaml:"database"`
	Consumer  consumer.Config  `yaml:"consumer"`
	Producer  producer.Config  `yaml:"producer"`
	Certer    certer.Config    `yaml:"certer"`
	Setter    setter.Config    `yaml:"setter"`
	Scheduler scheduler.Config `yaml:"scheduler"`
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
