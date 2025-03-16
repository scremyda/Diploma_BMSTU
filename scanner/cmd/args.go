package main

import "flag"

func ReadArgs() string {
	var configPath string
	flag.StringVar(&configPath, "c", "./config.yaml", "Path to YAML config file")

	flag.Parse()

	return configPath
}
