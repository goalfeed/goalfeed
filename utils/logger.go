package utils

import (
	"sync"

	"go.uber.org/zap"
)

var (
	logger *zap.SugaredLogger
	once   sync.Once
)

func GetLogger() *zap.SugaredLogger {
	once.Do(func() {
		zapLogger, _ := zap.NewProduction()
		logger = zapLogger.Sugar()
	})
	return logger
}
