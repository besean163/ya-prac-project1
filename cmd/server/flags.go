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
)

type ServerConfig struct {
	Endpoint      string
	StoreInterval int
	StoreFile     string
	Restore       bool
	BaseDNS       string
	HashKey       string
}

func NewConfig() ServerConfig {
	endpointFlag := ""
	storeIntervalFlag := 0
	storeFileFlag := ""
	restoreFlag := false
	baseDSNFlag := ""
	hashKeyFlag := ""
	flag.StringVar(&endpointFlag, "a", endpointDefault, "server endpoint")
	flag.IntVar(&storeIntervalFlag, "i", storeIntervalDefault, "store interval")
	flag.StringVar(&storeFileFlag, "f", storeFileDefault, "store file")
	flag.BoolVar(&restoreFlag, "r", restoreFlagDefault, "restore metrics")
	flag.StringVar(&baseDSNFlag, "d", baseDSNDefault, "data base dsn")
	flag.StringVar(&hashKeyFlag, "k", hashKeyDefault, "hash key")
	flag.Parse()

	if endpointEnv := os.Getenv("ADDRESS"); endpointEnv != "" {
		endpointFlag = endpointEnv
	}

	if storeIntervalEnv := os.Getenv("STORE_INTERVAL"); storeIntervalEnv != "" {
		i, err := strconv.Atoi(storeIntervalEnv)
		if err == nil {
			storeIntervalFlag = i
		}
	}

	if storeFileEnv := os.Getenv("FILE_STORAGE_PATH"); storeFileEnv != "" {
		storeFileFlag = storeFileEnv
	}

	if restoreEnv := os.Getenv("RESTORE"); restoreEnv != "" {
		restore, err := strconv.ParseBool(restoreEnv)
		if err == nil {
			restoreFlag = restore
		}
	}

	if baseDSNEnv := os.Getenv("DATABASE_DSN"); baseDSNEnv != "" {
		baseDSNFlag = baseDSNEnv
	}

	if hashKeyEnv := os.Getenv("KEY"); hashKeyEnv != "" {
		hashKeyFlag = hashKeyEnv
	}

	return ServerConfig{
		Endpoint:      endpointFlag,
		StoreInterval: storeIntervalFlag,
		StoreFile:     storeFileFlag,
		Restore:       restoreFlag,
		BaseDNS:       baseDSNFlag,
		HashKey:       hashKeyFlag,
	}
}
