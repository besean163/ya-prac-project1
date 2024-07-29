package main

import (
	"flag"
	"os"
	"strconv"
)

type ServerConfig struct {
	Endpoint      string
	StoreInterval int
	StoreFile     string
	Restore       bool
}

func NewConfig() ServerConfig {
	endpointFlag := ""
	storeIntervalFlag := 0
	storeFileFlag := ""
	restoreFlag := false
	flag.StringVar(&endpointFlag, "a", "localhost:8080", "server endpoint")
	flag.IntVar(&storeIntervalFlag, "i", 300, "store interval")
	flag.StringVar(&storeFileFlag, "f", "store_metrics", "store file")
	flag.BoolVar(&restoreFlag, "r", true, "restore metrics")
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

	return ServerConfig{
		Endpoint:      endpointFlag,
		StoreInterval: storeIntervalFlag,
		StoreFile:     storeFileFlag,
		Restore:       restoreFlag,
	}
}
