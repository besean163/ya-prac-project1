package main

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

// MockClient реализует HTTPClientInterface для тестирования
type MockClient struct {
	Response *http.Response
	Err      error
	Attempts int
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	m.Attempts++
	return m.Response, m.Err
}

func TestWorker_SuccessfulRequest(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	requestCh := make(chan *http.Request, 1)
	requestDone := make(chan struct{}, 1)

	mockResponse := &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}
	mockClient := &MockClient{Response: mockResponse}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	requestCh <- req

	go worker(ctx, requestCh, requestDone, mockClient)

	select {
	case <-requestDone:
		if mockClient.Attempts != 1 {
			t.Errorf("expected 1 attempt, got %d", mockClient.Attempts)
		}
	case <-time.After(1 * time.Second):
		t.Error("worker did not finish in time")
	}
}

func TestWorker_RetryOnError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	requestCh := make(chan *http.Request, 1)
	requestDone := make(chan struct{}, 1)

	mockClient := &MockClient{
		Err: errors.New("connection refused"),
	}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	requestCh <- req

	go worker(ctx, requestCh, requestDone, mockClient)

	select {
	case <-requestDone:
		if mockClient.Attempts != retryAttempts {
			t.Errorf("expected %d attempts, got %d", retryAttempts, mockClient.Attempts)
		}
	case <-time.After(6 * time.Second):
		t.Error("worker did not finish retries in time")
	}
}

func TestWorker_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	requestCh := make(chan *http.Request, 1)
	requestDone := make(chan struct{}, 1)

	mockClient := &MockClient{}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	requestCh <- req

	go worker(ctx, requestCh, requestDone, mockClient)

	cancel() // Отмена контекста

	select {
	case <-time.After(1 * time.Second):
		if mockClient.Attempts > 0 {
			t.Errorf("expected 0 attempts after cancellation, got %d", mockClient.Attempts)
		}
	case <-requestDone:
		// t.Error("worker should not have completed a request after cancellation")
	}
}

func TestWorker_IncreasingWaitTimes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	requestCh := make(chan *http.Request, 1)
	requestDone := make(chan struct{}, 1)

	mockClient := &MockClient{
		Err: errors.New("connection refused"),
	}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	requestCh <- req

	start := time.Now()
	go worker(ctx, requestCh, requestDone, mockClient)

	select {
	case <-requestDone:
		duration := time.Since(start)
		expected := waitSec + waitSecIncrement*(retryAttempts-1)
		if int(duration.Seconds()) < expected {
			t.Errorf("expected worker to wait at least %d seconds, got %v", expected, duration)
		}
	case <-time.After(10 * time.Second):
		t.Error("worker did not finish in time")
	}
}
