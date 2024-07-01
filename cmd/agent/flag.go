package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
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

	if envServerAddr := os.Getenv("ADDRESS"); envServerAddr != "" {
		serverEndpointFlag = envServerAddr
	}

	if envReportInv := os.Getenv("REPORT_INTERVAL"); envReportInv != "" {
		interval, err := strconv.Atoi(envReportInv)
		if err == nil {
			reportIntervalFlag = interval
		}
	}

	if envPoolInv := os.Getenv("POLL_INTERVAL"); envPoolInv != "" {
		interval, err := strconv.Atoi(envPoolInv)
		if err == nil {
			pollIntervalFlag = interval
		}
	}

	fmt.Println("Run with:")
	fmt.Println("Storage server endpoint: ", serverEndpointFlag)
	fmt.Println("Report interval (sec):", reportIntervalFlag)
	fmt.Println("Pool interval (sec):", pollIntervalFlag)
}
