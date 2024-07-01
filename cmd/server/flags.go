package main

import "flag"

var serverEndpointFlag string

func parseFlags() {
	flag.StringVar(&serverEndpointFlag, "a", "localhost:8080", "server endpoint")
	flag.Parse()
}
