package main

import (
	"log"
	"net/http"
	"strings"
	"time"
	"ya-prac-project1/internal/logger"
	"ya-prac-project1/internal/services"
	"ya-prac-project1/internal/storage/inmemstorage"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	retryAttempts    = 3
	waitSec          = 1
	waitSecIncrement = 2
)

func main() {
	c := NewConfig()
	errGroup := new(errgroup.Group)
	requestCh := make(chan *http.Request)

	storage := inmemstorage.NewStorage()
	service := services.NewRuntimeService(storage)

	for count := 0; count < c.RateLimit; count++ {
		errGroup.Go(func() error {
			worker(requestCh)
			return nil
		})
	}

	service.Run(errGroup, c.PoolInterval)

	errGroup.Go(func() error {
		for {
			time.Sleep(time.Duration(c.ReportInterval) * time.Second)
			service.RunSendRequest(requestCh, c.Endpoint, c.HashKey)
		}
	})

	// errGroup.Go(func() error {
	// 	for {
	// 		service.SendMetrics(c.Endpoint, c.HashKey)
	// 		time.Sleep(time.Duration(c.ReportInterval) * time.Second)
	// 	}
	// })

	if err := errGroup.Wait(); err != nil {
		logger.Get().Debug("agent error", zap.String("error", err.Error()))
	}
}

func worker(requestCh chan *http.Request) {
	for {
		req := <-requestCh

		client := http.Client{}
		var err error
		retry := getRetryFunc(retryAttempts, waitSec, waitSecIncrement)
		var response *http.Response
		for retry(err) {
			response, err = client.Do(req)
		}

		if err != nil {
			log.Printf("call error. Error: %s\n", err)
			continue
		}

		if response.StatusCode != http.StatusOK {
			log.Println("Error write metrics")
			log.Println("Path:", req.URL.Path)
			log.Println("Code:", response.StatusCode)
		}
		response.Body.Close()
	}
}

// возвращает функцию которая следит за количеством повторов и определяет их надобность
// работает за счет замыкания, т.е. передаем в параметры создающей функции количество попыток и каждую следующую заддержку и эти параметры используем при каждом вызове функции
func getRetryFunc(attempts, secDelta, waitDelta int) func(err error) bool {
	attempt := 0
	return func(err error) bool {
		attempt++

		// первый запуск
		if attempt == 1 && attempt <= attempts {
			return true
		}

		if err == nil {
			return false
		}

		if strings.Contains(err.Error(), "connection refused") {
			time.Sleep(time.Duration(secDelta) * time.Second)
			secDelta += waitDelta
			return attempt <= attempts
		}

		// если дошли сюда, то попытки закончились
		return false
	}
}
