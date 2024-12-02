package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

func worker(ctx context.Context, requestCh chan *http.Request, requestDone chan struct{}, client HTTPClientInterface) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("worker stopped")
			return
		case req := <-requestCh:
			// client := http.Client{}
			var err error

			response, err := client.Do(req)
			if response != nil && response.Body != nil {
				response.Body.Close()
			}
			attempt := 1
			secDelta := waitSec
			for response == nil && err != nil && needRetry(err) && attempt <= retryAttempts {
				log.Printf("get error, need try again, wait %d sec", secDelta)
				stop := false
				ticker := time.NewTicker(time.Duration(secDelta) * time.Second)
				select {
				case <-ticker.C:
					response, err = client.Do(req)
					if response != nil {
						response.Body.Close()
					}
				case <-ctx.Done():
					stop = true
				}

				if stop {
					break
				}

				attempt++
				secDelta += waitSecIncrement
			}

			if response != nil {
				if response.StatusCode != http.StatusOK {
					log.Println("Error write metrics")
					log.Println("Path:", req.URL.Path)
					log.Println("Code:", response.StatusCode)
				}
			}

			if err != nil {
				log.Printf("call error. Error: %s\n", err)
			}

			requestDone <- struct{}{}
		}

	}

}
