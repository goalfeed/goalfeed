package config

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetString(t *testing.T) {
	// Reset viper before test
	viper.Reset()
	
	// Set a test value
	viper.Set("test.key", "test value")
	
	result := GetString("test.key")
	assert.Equal(t, "test value", result)
	
	// Test non-existent key
	result = GetString("non.existent.key")
	assert.Equal(t, "", result)
}

func TestGetStringSlice(t *testing.T) {
	// Reset viper before test
	viper.Reset()
	
	// Set a test slice value
	testSlice := []string{"value1", "value2", "value3"}
	viper.Set("test.slice", testSlice)
	
	result := GetStringSlice("test.slice")
	assert.Equal(t, testSlice, result)
	
	// Test non-existent key
	result = GetStringSlice("non.existent.slice")
	assert.Empty(t, result)
}

func TestEnvironmentVariableOverride(t *testing.T) {
	// Reset viper before test
	viper.Reset()
	
	// Set an environment variable
	os.Setenv("GOALFEED_TEST_ENV", "env_value")
	defer os.Unsetenv("GOALFEED_TEST_ENV")
	
	// Reinitialize viper to pick up env vars
	viper.SetEnvPrefix("GOALFEED")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(testReplacer())
	
	result := GetString("test.env")
	assert.Equal(t, "env_value", result)
}

func TestConfigInitialization(t *testing.T) {
	// This test verifies that the init function doesn't panic
	// and sets up viper correctly
	assert.NotPanics(t, func() {
		// The init function should have already run
		// Just verify viper can read values without panicking
		_ = GetString("any.key")
		_ = GetStringSlice("any.slice")
	})
}

// Helper function to create a replacer for testing
func testReplacer() *strings.Replacer {
	return strings.NewReplacer(".", "_")
}