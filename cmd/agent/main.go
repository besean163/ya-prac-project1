package main

import (
	"log"
	"time"
	"ya-prac-project1/internal/inmem"
	"ya-prac-project1/internal/logger"
	"ya-prac-project1/internal/services"
)

func main() {
	if err := logger.Set(); err != nil {
		log.Fatalf(err.Error())
	}

	c := NewConfig()

	storage := inmem.NewStorage()
	service := services.NewRuntimeService(&storage)

	go func() {
		for {
			err := service.UpdateMetrics()
			if err != nil {
				logger.Get().Info("update metric error")
			}
			time.Sleep(time.Duration(c.PoolInterval) * time.Second)
		}
	}()

	for {
		time.Sleep(time.Duration(c.ReportInterval) * time.Second)
		service.SendMetrics(c.Endpoint)
	}
}
