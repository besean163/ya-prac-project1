package main

import (
	"flag"
	"os"
)

var serverEndpointFlag string

func parseFlags() {
	flag.StringVar(&serverEndpointFlag, "a", "localhost:8080", "server endpoint")
	flag.Parse()

	if envServerAddr := os.Getenv("ADDRESS"); envServerAddr != "" {
		serverEndpointFlag = envServerAddr
	}
}
