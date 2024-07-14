package main

import (
	"time"
	"ya-prac-project1/internal/inmem"
	"ya-prac-project1/internal/services"
)

func main() {
	c := NewConfig()

	storage := inmem.NewStorage()
	service := services.NewRuntimeService(&storage)

	go func() {
		for {
			service.UpdateMetrics()
			time.Sleep(time.Duration(c.PoolInterval) * time.Second)
		}
	}()

	for {
		time.Sleep(time.Duration(c.ReportInterval) * time.Second)
		service.SendMetrics(c.Endpoint)
	}
}
