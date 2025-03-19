package main

import (
	"diploma/scanner/analyzer"
	"diploma/scanner/scheduler"
	"diploma/scanner/scraper"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	defaultRangeFactor float64 = 0.1
)

type Config struct {
	External []struct {
		ScraperConf   scraper.Conf   `yaml:"scraper_conf"`
		AnalyzerConf  analyzer.Conf  `yaml:"analyzer_conf"`
		SchedulerConf scheduler.Conf `yaml:"scheduler_conf"`
	} `yaml:"external"`
	Internal InternalConf `yaml:"internal"`
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
	for i, cfg := range conf.External {
		if cfg.SchedulerConf.Interval < conf.Internal.MinScrapeInterval {
			log.Printf("Interval %s for %s is less than min_scrape_interval (%s); setting to min_scrape_interval",
				cfg.SchedulerConf.Interval, cfg.ScraperConf.Target, conf.Internal.MinScrapeInterval)

			conf.External[i].SchedulerConf.Interval = conf.Internal.MinScrapeInterval
		}

		if cfg.SchedulerConf.RangeFactor <= 0 {
			log.Printf("RangeFactor %f for %s is invalid; setting to default RangeFactor (%f)",
				cfg.SchedulerConf.RangeFactor, cfg.ScraperConf.Target, defaultRangeFactor)

			conf.External[i].SchedulerConf.RangeFactor = defaultRangeFactor
		}
	}
}
