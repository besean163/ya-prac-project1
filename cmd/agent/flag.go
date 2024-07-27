package main

import (
	"flag"
	"os"
	"strconv"
)

type AgentConfig struct {
	Endpoint       string
	ReportInterval int
	PoolInterval   int
}

const (
	defaultPoolInterval   = 1
	defaultReportInterval = 2
)

func NewConfig() AgentConfig {
	config := AgentConfig{}

	flag.StringVar(&config.Endpoint, "a", "localhost:8080", "server endpoint")
	flag.IntVar(&config.ReportInterval, "r", defaultReportInterval, "report interval sec")
	flag.IntVar(&config.PoolInterval, "p", defaultPoolInterval, "metrics pool interval sec")
	flag.Parse()

	if endpointEnv := os.Getenv("ADDRESS"); endpointEnv != "" {
		config.Endpoint = endpointEnv
	}

	if reportIntervalEnv := os.Getenv("REPORT_INTERVAL"); reportIntervalEnv != "" {
		interval, err := strconv.Atoi(reportIntervalEnv)
		if err == nil {
			config.ReportInterval = interval
		}
	}

	if poolIntervalEnv := os.Getenv("POLL_INTERVAL"); poolIntervalEnv != "" {
		interval, err := strconv.Atoi(poolIntervalEnv)
		if err == nil {
			config.PoolInterval = interval
		}
	}

	if config.PoolInterval == 0 {
		config.PoolInterval = defaultPoolInterval
	}

	if config.ReportInterval == 0 {
		config.ReportInterval = defaultReportInterval
	}

	return config
}
