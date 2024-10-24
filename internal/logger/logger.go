// Package logger предоставляет доступ к глобальному логгеру
package logger

import "go.uber.org/zap"

var state *zap.Logger

// Set установить глобальную переменную логгера
func Set() error {
	var err error
	state, err = zap.NewProduction()
	return err
}

// Get получить глобальную переменную логгера
func Get() *zap.Logger {
	return state
}
