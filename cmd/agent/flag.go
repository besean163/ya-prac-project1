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
	rateLimitDefault      = 1
)

type AgentConfig struct {
	Endpoint       string
	ReportInterval int
	PoolInterval   int
	HashKey        string
	RateLimit      int
}

func NewConfig() AgentConfig {
	config := AgentConfig{}

	flag.StringVar(&config.Endpoint, "a", endpointDefault, "server endpoint")
	flag.IntVar(&config.ReportInterval, "r", reportIntervalDefault, "report interval sec")
	flag.IntVar(&config.PoolInterval, "p", poolIntervalDefault, "metrics pool interval sec")
	flag.StringVar(&config.HashKey, "k", hashKeyDefault, "hash key")
	flag.IntVar(&config.RateLimit, "l", rateLimitDefault, "rate limit")
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

	if rateLimitEnv := os.Getenv("RATE_LIMIT"); rateLimitEnv != "" {
		rateLimit, err := strconv.Atoi(rateLimitEnv)
		if err == nil {
			config.RateLimit = rateLimit
		}
	}

	return config
}
