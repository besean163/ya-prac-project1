package main

import (
	"log"
	"time"
	"ya-prac-project1/internal/inmem"
	"ya-prac-project1/internal/services"
)

func main() {
	c := NewConfig()

	storage, err := inmem.NewStorage("", false, 0)
	if err != nil {
		log.Fatalf(err.Error())
	}

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
