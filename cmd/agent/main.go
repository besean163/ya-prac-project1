package main

import (
	"time"
	"ya-prac-project1/internal/services"
	"ya-prac-project1/internal/storage/inmemstorage"
)

func main() {
	c := NewConfig()

	storage := inmemstorage.NewStorage()
	service := services.NewRuntimeService(storage)

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
