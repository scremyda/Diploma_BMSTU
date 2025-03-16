package main

import (
	"diploma/scanner/scraper"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"time"
)

type Config struct {
	Scraper             []scraper.Conf `yaml:"scraper"`
	ScraperInternalConf InternalConf   `yaml:"scraper_internal"`
}

type InternalConf struct {
	PoolSize          int           `yaml:"pool_size"`
	MinScrapeInterval time.Duration `yaml:"min_scrape_interval"`
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

func (conf *Config) AdjustScrapeIntervals() {
	for i, cfg := range conf.Scraper {
		if cfg.Interval < conf.ScraperInternalConf.MinScrapeInterval {
			log.Printf("Interval %s for %s is less than min_scrape_interval (%s); setting to min_scrape_interval",
				cfg.Interval, cfg.Target, conf.ScraperInternalConf.MinScrapeInterval)
			conf.Scraper[i].Interval = conf.ScraperInternalConf.MinScrapeInterval
		}
	}
}
