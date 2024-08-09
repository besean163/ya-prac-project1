package main

import (
	"flag"
	"os"
	"strconv"
)

const (
	endpointDefault       = "localhost:8080"
	reportIntervalDefault = 2
	poolIntervalDefault   = 1
)

type AgentConfig struct {
	Endpoint       string
	ReportInterval int
	PoolInterval   int
}

func NewConfig() AgentConfig {
	config := AgentConfig{}

	flag.StringVar(&config.Endpoint, "a", endpointDefault, "server endpoint")
	flag.IntVar(&config.ReportInterval, "r", reportIntervalDefault, "report interval sec")
	flag.IntVar(&config.PoolInterval, "p", poolIntervalDefault, "metrics pool interval sec")
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

	return config
}
