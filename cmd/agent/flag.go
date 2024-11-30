package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
)

const (
	endpointDefault       = "localhost:8080"
	reportIntervalDefault = 2
	poolIntervalDefault   = 1
	hashKeyDefault        = ""
	rateLimitDefault      = 1
	profilerDefault       = ""
	cryptoKeyDefault      = ""
)

type AgentConfig struct {
	Endpoint       string `json:"address"`
	HashKey        string
	Profiler       string
	ReportInterval int `json:"report_interval"`
	PoolInterval   int `json:"poll_interval"`
	RateLimit      int
	CryptoKey      string `json:"crypto_keys"`
}

func NewDefaultConfig() AgentConfig {
	c := AgentConfig{
		Endpoint:       endpointDefault,
		HashKey:        hashKeyDefault,
		Profiler:       profilerDefault,
		ReportInterval: reportIntervalDefault,
		PoolInterval:   poolIntervalDefault,
		RateLimit:      rateLimitDefault,
		CryptoKey:      cryptoKeyDefault,
	}
	return c
}

func NewConfig() AgentConfig {
	configPath := ""
	flag.StringVar(&configPath, "c", getEnv("CONFIG", ""), "config path")

	defaultConfig := NewDefaultConfig()
	config, err := loadConfigFromFile(configPath)
	if err != nil {
		fmt.Println("config file read error", err)
		config = &defaultConfig
	}

	flag.StringVar(&config.Endpoint, "a", config.Endpoint, "server endpoint")
	flag.IntVar(&config.ReportInterval, "r", config.ReportInterval, "report interval sec")
	flag.IntVar(&config.PoolInterval, "p", config.PoolInterval, "metrics pool interval sec")
	flag.StringVar(&config.HashKey, "k", config.HashKey, "hash key")
	flag.IntVar(&config.RateLimit, "l", config.RateLimit, "rate limit")
	flag.StringVar(&config.Profiler, "profile", config.Profiler, "profiler port")
	flag.StringVar(&config.CryptoKey, "crypto-key", config.CryptoKey, "crypto key")

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

	if cryptoKeyEnv := os.Getenv("CRYPTO_KEY"); cryptoKeyEnv != "" {
		config.CryptoKey = cryptoKeyEnv
	}

	return *config
}

func loadConfigFromFile(filename string) (*AgentConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &AgentConfig{}
	err = json.NewDecoder(file).Decode(config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func getEnv(key string, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
