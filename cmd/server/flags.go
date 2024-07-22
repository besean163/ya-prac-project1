package main

import (
	"flag"
	"os"
)

type ServerConfig struct {
	Endpoint string
}

func NewConfig() ServerConfig {
	endpointFlag := ""
	flag.StringVar(&endpointFlag, "a", "localhost:8080", "server endpoint")
	flag.Parse()

	if endpointEnv := os.Getenv("ADDRESS"); endpointEnv != "" {
		endpointFlag = endpointEnv
	}

	return ServerConfig{
		Endpoint: endpointFlag,
	}
}
