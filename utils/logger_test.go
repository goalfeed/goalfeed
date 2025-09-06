package utils

import (
	"testing"

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