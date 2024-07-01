package main

import (
	"flag"
	"fmt"
)

var serverEndpointFlag string
var reportIntervalFlag int
var pollIntervalFlag int

const (
	poolInterval   = 1
	reportInterval = 2
)

func parseFlags() {
	flag.StringVar(&serverEndpointFlag, "a", "localhost:8080", "server endpoint")
	flag.IntVar(&reportIntervalFlag, "r", reportInterval, "report interval sec")
	flag.IntVar(&pollIntervalFlag, "p", poolInterval, "metrics pool interval sec")
	flag.Parse()

	fmt.Println("Run with:")
	fmt.Println("Storage server endpoint: ", serverEndpointFlag)
	fmt.Println("Report interval (sec):", reportIntervalFlag)
	fmt.Println("Pool interval (sec):", pollIntervalFlag)
}
