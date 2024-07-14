package main

import (
	"time"
	"ya-prac-project1/internal/inmem"
	"ya-prac-project1/internal/services"
)

func main() {
	parseFlags()

	storage := inmem.MemStorage{}
	service := services.NewRuntimeService(&storage)

	go func() {
		for {
			service.UpdateMetrics()
			time.Sleep(time.Duration(pollIntervalFlag) * time.Second)
		}
	}()

	for {
		time.Sleep(time.Duration(reportIntervalFlag) * time.Second)
		service.SendMetrics(serverEndpointFlag)
	}
}
