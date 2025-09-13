package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGetLogger(t *testing.T) {
	logger := GetLogger()

	// Verify that logger is not nil
	assert.NotNil(t, logger)

	// Verify that it's a SugaredLogger
	assert.IsType(t, &zap.SugaredLogger{}, logger)

	// Test that we can use the logger without errors
	assert.NotPanics(t, func() {
		logger.Info("Test log message")
	})
}

func TestGetLogger_Singleton(t *testing.T) {
	// Test that multiple calls return the same logger instance
	logger1 := GetLogger()
	logger2 := GetLogger()

	// Should return the same instance (same memory address)
	assert.Equal(t, logger1, logger2)
}

func TestGetLogger_StructuredLogging(t *testing.T) {
	logger := GetLogger()

	// Test structured logging
	assert.NotPanics(t, func() {
		logger.Infow("test message",
			"key", "value",
			"number", 42,
			"flag", true,
		)
	})
}

func TestGetLogger_ErrorLogging(t *testing.T) {
	logger := GetLogger()

	// Test error logging
	assert.NotPanics(t, func() {
		logger.Errorw("test error message",
			"error", assert.AnError,
			"context", "test",
		)
	})
}

func TestGetLogger_WarnLogging(t *testing.T) {
	logger := GetLogger()

	// Test warning logging
	assert.NotPanics(t, func() {
		logger.Warnw("test warning message",
			"warning", "test warning",
		)
	})
}

func TestGetLogger_DebugLogging(t *testing.T) {
	logger := GetLogger()

	// Test debug logging
	assert.NotPanics(t, func() {
		logger.Debugw("test debug message",
			"debug", "test debug",
		)
	})
}

func TestGetLogger_WithFields(t *testing.T) {
	logger := GetLogger()

	// Test logging with fields
	assert.NotPanics(t, func() {
		logger.With(
			"service", "test",
			"timestamp", time.Now(),
		).Infow("test message with fields")
	})
}

func TestGetLogger_ConcurrentAccess(t *testing.T) {
	logger := GetLogger()

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			defer func() { done <- true }()
			assert.NotPanics(t, func() {
				logger.Infow("concurrent test message",
					"goroutine", i,
				)
			})
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestGetLogger_FormatLogging(t *testing.T) {
	logger := GetLogger()

	// Test format logging
	assert.NotPanics(t, func() {
		logger.Infof("test format message: %s", "test value")
		logger.Errorf("test error format: %d", 42)
		logger.Warnf("test warning format: %v", true)
		logger.Debugf("test debug format: %s", "debug value")
	})
}
