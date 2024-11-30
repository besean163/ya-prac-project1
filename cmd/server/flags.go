package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
)

const (
	endpointDefault      = "localhost:8080"
	storeIntervalDefault = 300
	storeFileDefault     = "store_metrics"
	restoreFlagDefault   = true
	baseDSNDefault       = ""
	hashKeyDefault       = ""
	profilerDefault      = ""
	cryptoKeyDefault     = ""
)

type ServerConfig struct {
	Endpoint      string
	StoreFile     string
	BaseDNS       string
	HashKey       string
	Profiler      string
	Restore       bool
	StoreInterval int
	CryptoKey     string
}

func NewDefaultConfig() ServerConfig {
	c := ServerConfig{
		Endpoint:      endpointDefault,
		StoreFile:     storeFileDefault,
		BaseDNS:       baseDSNDefault,
		HashKey:       hashKeyDefault,
		Profiler:      profilerDefault,
		Restore:       restoreFlagDefault,
		StoreInterval: storeIntervalDefault,
		CryptoKey:     cryptoKeyDefault,
	}
	return c
}

func NewConfig() ServerConfig {

	configPath := ""
	flag.StringVar(&configPath, "c", getEnv("CONFIG", ""), "config path")

	defaultConfig := NewDefaultConfig()
	config, err := loadConfigFromFile(configPath)
	if err != nil {
		fmt.Println("config file read error", err)
		config = &defaultConfig
	}

	flag.StringVar(&config.Endpoint, "a", config.Endpoint, "server endpoint")
	flag.IntVar(&config.StoreInterval, "i", config.StoreInterval, "store interval")
	flag.StringVar(&config.StoreFile, "f", config.StoreFile, "store file")
	flag.BoolVar(&config.Restore, "r", config.Restore, "restore metrics")
	flag.StringVar(&config.BaseDNS, "d", config.BaseDNS, "data base dsn")
	flag.StringVar(&config.HashKey, "k", config.HashKey, "hash key")
	flag.StringVar(&config.Profiler, "p", config.Profiler, "profiler port")
	flag.StringVar(&config.CryptoKey, "crypto-key", config.CryptoKey, "crypto key")
	flag.Parse()

	if endpointEnv := os.Getenv("ADDRESS"); endpointEnv != "" {
		config.Endpoint = endpointEnv
	}

	if storeIntervalEnv := os.Getenv("STORE_INTERVAL"); storeIntervalEnv != "" {
		i, err := strconv.Atoi(storeIntervalEnv)
		if err == nil {
			config.StoreInterval = i
		}
	}

	if storeFileEnv := os.Getenv("FILE_STORAGE_PATH"); storeFileEnv != "" {
		config.StoreFile = storeFileEnv
	}

	if restoreEnv := os.Getenv("RESTORE"); restoreEnv != "" {
		restore, err := strconv.ParseBool(restoreEnv)
		if err == nil {
			config.Restore = restore
		}
	}

	if baseDSNEnv := os.Getenv("DATABASE_DSN"); baseDSNEnv != "" {
		config.BaseDNS = baseDSNEnv
	}

	if hashKeyEnv := os.Getenv("KEY"); hashKeyEnv != "" {
		config.HashKey = hashKeyEnv
	}

	if cryptoKeyEnv := os.Getenv("CRYPTO_KEY"); cryptoKeyEnv != "" {
		config.CryptoKey = cryptoKeyEnv
	}

	return *config
}

func loadConfigFromFile(filename string) (*ServerConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &ServerConfig{}
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
