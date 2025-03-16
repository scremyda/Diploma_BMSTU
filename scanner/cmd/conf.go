package main

import (
	"diploma/scanner/scraper"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Scraper             []scraper.Conf       `yaml:"scraper"`
	ScraperInternalConf scraper.InternalConf `yaml:"scraper_internal"`
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
