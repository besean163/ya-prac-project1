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
	hashKeyDefault        = ""
)

type AgentConfig struct {
	Endpoint       string
	ReportInterval int
	PoolInterval   int
	HashKey        string
}

func NewConfig() AgentConfig {
	config := AgentConfig{}

	flag.StringVar(&config.Endpoint, "a", endpointDefault, "server endpoint")
	flag.IntVar(&config.ReportInterval, "r", reportIntervalDefault, "report interval sec")
	flag.IntVar(&config.PoolInterval, "p", poolIntervalDefault, "metrics pool interval sec")
	flag.StringVar(&config.HashKey, "k", hashKeyDefault, "hash key")
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

	if hashKeyEnv := os.Getenv("KEY"); hashKeyEnv != "" {
		config.HashKey = hashKeyEnv
	}

	return config
}
