package main

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestTeamIsMonitoredByLeague(t *testing.T) {
	// Reset viper before each test
	defer viper.Reset()

	// Test with command-line arguments
	viper.Set("watch.nhl", []string{"WPG"})
	viper.Set("watch.mlb", []string{"TOR"})
	assert.True(t, teamIsMonitoredByLeague("WPG", "nhl"), "Expected TOR to be monitored for NHL based on command-line arguments")

	// Test with different cases
	assert.True(t, teamIsMonitoredByLeague("wpg", "nhl"), "Expected wpg (in lowercase) to be monitored for NHL")
	assert.True(t, teamIsMonitoredByLeague("WpG", "nhl"), "Expected WpG (in mixed case) to be monitored for NHL")

	// Test with configuration
	viper.SetConfigName("config.example")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		t.Fatalf("Error reading config: %v", err)
	}
	assert.True(t, teamIsMonitoredByLeague("TOR", "mlb"), "Expected TOR to be monitored for NHL based on environment variable")
	assert.True(t, teamIsMonitoredByLeague("WPG", "nhl"), "Expected WPG to be monitored for NHL based on environment variable")
	assert.False(t, teamIsMonitoredByLeague("TOR", "nhl"), "Expected TOR to be monitored for NHL based on environment variable")

	// Test with environment variables
	os.Setenv("WATCH_NHL", "WPG")
	os.Setenv("WATCH_MLB", "TOR")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	assert.True(t, teamIsMonitoredByLeague("TOR", "mlb"), "Expected TOR to be monitored for NHL based on environment variable")
	assert.True(t, teamIsMonitoredByLeague("WPG", "nhl"), "Expected WPG to be monitored for NHL based on environment variable")
	assert.False(t, teamIsMonitoredByLeague("TOR", "nhl"), "Expected TOR to be monitored for NHL based on environment variable")
}
