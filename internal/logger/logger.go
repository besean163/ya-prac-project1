package logger

import "go.uber.org/zap"

var state *zap.Logger

func Set() error {
	var err error
	state, err = zap.NewProduction()
	return err
}

func Get() *zap.Logger {
	return state
}
