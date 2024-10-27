package main

import (
	"flag"
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
)

type ServerConfig struct {
	Endpoint      string
	StoreFile     string
	BaseDNS       string
	HashKey       string
	Profiler      string
	Restore       bool
	StoreInterval int
}

func NewConfig() ServerConfig {
	c := ServerConfig{}

	flag.StringVar(&c.Endpoint, "a", endpointDefault, "server endpoint")
	flag.IntVar(&c.StoreInterval, "i", storeIntervalDefault, "store interval")
	flag.StringVar(&c.StoreFile, "f", storeFileDefault, "store file")
	flag.BoolVar(&c.Restore, "r", restoreFlagDefault, "restore metrics")
	flag.StringVar(&c.BaseDNS, "d", baseDSNDefault, "data base dsn")
	flag.StringVar(&c.HashKey, "k", hashKeyDefault, "hash key")
	flag.StringVar(&c.Profiler, "p", profilerDefault, "profiler port")
	flag.Parse()

	if endpointEnv := os.Getenv("ADDRESS"); endpointEnv != "" {
		c.Endpoint = endpointEnv
	}

	if storeIntervalEnv := os.Getenv("STORE_INTERVAL"); storeIntervalEnv != "" {
		i, err := strconv.Atoi(storeIntervalEnv)
		if err == nil {
			c.StoreInterval = i
		}
	}

	if storeFileEnv := os.Getenv("FILE_STORAGE_PATH"); storeFileEnv != "" {
		c.StoreFile = storeFileEnv
	}

	if restoreEnv := os.Getenv("RESTORE"); restoreEnv != "" {
		restore, err := strconv.ParseBool(restoreEnv)
		if err == nil {
			c.Restore = restore
		}
	}

	if baseDSNEnv := os.Getenv("DATABASE_DSN"); baseDSNEnv != "" {
		c.BaseDNS = baseDSNEnv
	}

	if hashKeyEnv := os.Getenv("KEY"); hashKeyEnv != "" {
		c.HashKey = hashKeyEnv
	}

	return c
}
