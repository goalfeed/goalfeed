package main

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"goalfeed/models"
	"goalfeed/targets/memoryStore"
)

// Test Helpers
func setupTest(t *testing.T) {
	viper.Reset()
	memoryStore.SetActiveGameKeys([]string{})
}

func createTestGame(leagueId models.League, homeTeam, awayTeam string) models.Game {
	return models.Game{
		LeagueId: leagueId,
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team:  models.Team{TeamCode: homeTeam},
				Score: 0,
			},
			Away: models.TeamState{
				Team:  models.Team{TeamCode: awayTeam},
				Score: 0,
			},
			Status: models.GameStatus(models.StatusActive),
		},
	}
}

// Test teamIsMonitoredByLeague function
func TestTeamIsMonitoredByLeague(t *testing.T) {
	// Reset viper before each test
	defer viper.Reset()

	// Test with command-line arguments
	viper.Set("watch.nhl", []string{"WPG"})
	viper.Set("watch.mlb", []string{"TOR"})
	assert.True(t, teamIsMonitoredByLeague("WPG", "nhl"), "Expected WPG to be monitored for NHL based on command-line arguments")

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
	assert.True(t, teamIsMonitoredByLeague("TOR", "mlb"), "Expected TOR to be monitored for MLB based on environment variable")
	assert.True(t, teamIsMonitoredByLeague("WPG", "nhl"), "Expected WPG to be monitored for NHL based on environment variable")
	assert.False(t, teamIsMonitoredByLeague("TOR", "nhl"), "Expected TOR to NOT be monitored for NHL based on environment variable")

	// Test with environment variables
	os.Setenv("WATCH_NHL", "WPG")
	os.Setenv("WATCH_MLB", "TOR")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	assert.True(t, teamIsMonitoredByLeague("TOR", "mlb"), "Expected TOR to be monitored for MLB based on environment variable")
	assert.True(t, teamIsMonitoredByLeague("WPG", "nhl"), "Expected WPG to be monitored for NHL based on environment variable")
	assert.False(t, teamIsMonitoredByLeague("TOR", "nhl"), "Expected TOR to NOT be monitored for NHL based on environment variable")

	// Test with wildcard "*"
	viper.Reset()
	viper.Set("watch.nhl", []string{"*"})
	assert.True(t, teamIsMonitoredByLeague("ANY", "nhl"), "Expected any team to be monitored when wildcard is used")
	assert.True(t, teamIsMonitoredByLeague("TOR", "nhl"), "Expected TOR to be monitored when wildcard is used")
	assert.True(t, teamIsMonitoredByLeague("WPG", "nhl"), "Expected WPG to be monitored when wildcard is used")
}

func TestInitialize(t *testing.T) {
	setupTest(t)

	// Test initialization of league services
	initialize()

	// We can't directly test the leagueServices map since it's not exported
	// But we can test that initialize() runs without error
}

func TestGameIsMonitored(t *testing.T) {
	setupTest(t)

	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")

	// Test when game is not monitored
	assert.False(t, gameIsMonitored(game))

	// Test when game is monitored
	memoryStore.AppendActiveGame(game)
	assert.True(t, gameIsMonitored(game))
}

func TestSendTestGoal(t *testing.T) {
	setupTest(t)

	// Test with test-goals disabled (safer for testing)
	viper.Set("test-goals", false)
	assert.NotPanics(t, func() {
		sendTestGoal()
	})
}

func TestPublishSchedules(t *testing.T) {
	setupTest(t)

	// Test that publishSchedules runs without error
	assert.NotPanics(t, func() {
		publishSchedules()
	})
}
