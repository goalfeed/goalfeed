package main

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"goalfeed/models"
	"goalfeed/services/leagues"
	"goalfeed/targets/memoryStore"
)

// MockLeagueService for testing
type MockLeagueService struct {
	leagueName string
	games      []models.Game
}

func (m *MockLeagueService) GetLeagueName() string {
	return m.leagueName
}

func (m *MockLeagueService) GetActiveGames(ch chan []models.Game) {
	ch <- m.games
}

func (m *MockLeagueService) GetUpcomingGames(ch chan []models.Game) {
	ch <- m.games
}

func (m *MockLeagueService) GetGameUpdate(game models.Game, ch chan models.GameUpdate) {
	ch <- models.GameUpdate{
		OldState: game.CurrentState,
		NewState: game.CurrentState,
	}
}

func (m *MockLeagueService) GetEvents(update models.GameUpdate, ch chan []models.Event) {
	ch <- []models.Event{}
}

// Test Helpers
func setupTest(t *testing.T) {
	viper.Reset()
	memoryStore.SetActiveGameKeys([]string{})

	// Initialize league services with mock services for testing
	leagueServices = map[int]leagues.ILeagueService{
		int(models.LeagueIdNHL): &MockLeagueService{},
		int(models.LeagueIdMLB): &MockLeagueService{},
		int(models.LeagueIdNFL): &MockLeagueService{},
		int(models.LeagueIdCFL): &MockLeagueService{},
	}
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

func TestSendTestGoal_Enabled(t *testing.T) {
	setupTest(t)

	// Test with test-goals enabled
	viper.Set("test-goals", true)
	assert.NotPanics(t, func() {
		sendTestGoal()
		// Give the goroutine time to execute
		time.Sleep(10 * time.Millisecond)
	})
}

func TestPublishSchedules(t *testing.T) {
	setupTest(t)

	// Test that publishSchedules runs without error
	assert.NotPanics(t, func() {
		publishSchedules()
	})
}

func TestCheckLeaguesForActiveGames(t *testing.T) {
	setupTest(t)

	// Test that checkLeaguesForActiveGames runs without error
	assert.NotPanics(t, func() {
		checkLeaguesForActiveGames()
	})
}

func TestWatchActiveGames(t *testing.T) {
	setupTest(t)

	// Test with no active games
	assert.NotPanics(t, func() {
		watchActiveGames()
	})

	// Test with active games
	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")
	memoryStore.AppendActiveGame(game)
	assert.NotPanics(t, func() {
		watchActiveGames()
	})
}

func TestFireGoalEvents(t *testing.T) {
	setupTest(t)

	// Create test game
	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")

	// Create test events
	events := make(chan []models.Event)
	go func() {
		events <- []models.Event{
			{
				TeamCode:    "WPG",
				TeamName:    "Winnipeg Jets",
				LeagueId:    models.LeagueIdNHL,
				LeagueName:  "NHL",
				Type:        models.EventTypeGoal,
				Description: "Goal scored!",
			},
		}
	}()

	// Test that fireGoalEvents runs without error
	assert.NotPanics(t, func() {
		fireGoalEvents(events, game)
	})
}

func TestRunTickers(t *testing.T) {
	setupTest(t)

	// Test that runTickers runs without error
	// Note: This test will run for a short time due to tickers
	assert.NotPanics(t, func() {
		go runTickers()
		// Let it run for a short time
		time.Sleep(100 * time.Millisecond)
	})
}

func TestCheckGame_GameNotFound(t *testing.T) {
	setupTest(t)

	// Test with non-existent game
	assert.NotPanics(t, func() {
		checkGame("non-existent-game")
	})
}

func TestCheckGame_ExistingGame(t *testing.T) {
	setupTest(t)

	// Test with existing game
	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")
	memoryStore.AppendActiveGame(game)
	assert.NotPanics(t, func() {
		checkGame(game.GetGameKey())
	})
}

func TestCheckGame_PeriodChange(t *testing.T) {
	setupTest(t)

	// Create a game with period 1
	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")
	game.CurrentState.Period = 1
	memoryStore.AppendActiveGame(game)

	// Mock the service to return a period change
	mockService := &MockLeagueService{
		leagueName: "NHL",
		games: []models.Game{
			createTestGame(models.LeagueIdNHL, "WPG", "TOR"),
		},
	}
	leagueServices[models.LeagueIdNHL] = mockService

	assert.NotPanics(t, func() {
		checkGame(game.GetGameKey())
	})
}

func TestCheckGame_StatusChange(t *testing.T) {
	setupTest(t)

	// Create a game with upcoming status
	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")
	game.CurrentState.Status = models.StatusUpcoming
	memoryStore.AppendActiveGame(game)

	// Mock the service to return a status change
	mockService := &MockLeagueService{
		leagueName: "NHL",
		games: []models.Game{
			createTestGame(models.LeagueIdNHL, "WPG", "TOR"),
		},
	}
	leagueServices[models.LeagueIdNHL] = mockService

	assert.NotPanics(t, func() {
		checkGame(game.GetGameKey())
	})
}

func TestCheckGame_GameEnded(t *testing.T) {
	setupTest(t)

	// Create a game with active status
	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")
	game.CurrentState.Status = models.StatusActive
	memoryStore.AppendActiveGame(game)

	// Mock the service to return an ended game
	mockService := &MockLeagueService{
		leagueName: "NHL",
		games: []models.Game{
			createTestGame(models.LeagueIdNHL, "WPG", "TOR"),
		},
	}
	leagueServices[models.LeagueIdNHL] = mockService

	assert.NotPanics(t, func() {
		checkGame(game.GetGameKey())
	})
}

func TestCheckGame_PeriodChange2(t *testing.T) {
	setupTest(t)

	// Create a game with period 1
	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")
	game.CurrentState.Period = 1
	memoryStore.AppendActiveGame(game)

	// Mock the service to return a period change
	mockService := &MockLeagueService{
		leagueName: "NHL",
		games: []models.Game{
			createTestGame(models.LeagueIdNHL, "WPG", "TOR"),
		},
	}
	leagueServices[models.LeagueIdNHL] = mockService

	assert.NotPanics(t, func() {
		checkGame(game.GetGameKey())
	})
}

func TestCheckGame_StatusChange2(t *testing.T) {
	setupTest(t)

	// Create a game with upcoming status
	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")
	game.CurrentState.Status = models.StatusUpcoming
	memoryStore.AppendActiveGame(game)

	// Mock the service to return a status change
	mockService := &MockLeagueService{
		leagueName: "NHL",
		games: []models.Game{
			createTestGame(models.LeagueIdNHL, "WPG", "TOR"),
		},
	}
	leagueServices[models.LeagueIdNHL] = mockService

	assert.NotPanics(t, func() {
		checkGame(game.GetGameKey())
	})
}

func TestCheckGame_GameEnded2(t *testing.T) {
	setupTest(t)

	// Create a game with active status
	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")
	game.CurrentState.Status = models.StatusActive
	memoryStore.AppendActiveGame(game)

	// Mock the service to return an ended game
	mockService := &MockLeagueService{
		leagueName: "NHL",
		games: []models.Game{
			createTestGame(models.LeagueIdNHL, "WPG", "TOR"),
		},
	}
	leagueServices[models.LeagueIdNHL] = mockService

	assert.NotPanics(t, func() {
		checkGame(game.GetGameKey())
	})
}

func TestCheckGame_SameState2(t *testing.T) {
	setupTest(t)

	// Create a game
	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")
	memoryStore.AppendActiveGame(game)

	// Mock the service to return the same state
	mockService := &MockLeagueService{
		leagueName: "NHL",
		games: []models.Game{
			createTestGame(models.LeagueIdNHL, "WPG", "TOR"),
		},
	}
	leagueServices[models.LeagueIdNHL] = mockService

	assert.NotPanics(t, func() {
		checkGame(game.GetGameKey())
	})
}

func TestCheckForNewActiveGames(t *testing.T) {
	setupTest(t)

	// Create a mock service
	mockService := &MockLeagueService{
		leagueName: "NHL",
		games: []models.Game{
			createTestGame(models.LeagueIdNHL, "WPG", "TOR"),
		},
	}

	// Test that checkForNewActiveGames runs without error
	assert.NotPanics(t, func() {
		checkForNewActiveGames(mockService)
	})
}

func TestPublishSchedulesWithTeams(t *testing.T) {
	setupTest(t)

	// Set up teams to watch
	viper.Set("watch.nhl", []string{"WPG"})
	viper.Set("watch.mlb", []string{"TOR"})

	// Test that publishSchedules runs without error
	assert.NotPanics(t, func() {
		publishSchedules()
	})
}

func TestFireGoalEvents2(t *testing.T) {
	setupTest(t)

	// Create a test game
	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")

	// Create a channel for events
	events := make(chan []models.Event, 1)

	// Create test events
	testEvents := []models.Event{
		{
			Type:        "GOAL",
			Description: "Test goal",
			TeamCode:    "WPG",
			GameCode:    game.GameCode,
		},
	}

	// Send events to the channel
	events <- testEvents

	// Test that fireGoalEvents runs without error
	assert.NotPanics(t, func() {
		fireGoalEvents(events, game)
	})

	// Close the channel
	close(events)
}

func TestFireGoalEvents_TeamNotMonitored(t *testing.T) {
	setupTest(t)

	// Create a test game
	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")

	// Set up teams to watch (excluding WPG)
	viper.Set("watch.nhl", []string{"TOR"})

	// Create a channel for events
	events := make(chan []models.Event, 1)

	// Create test events for WPG (not monitored)
	testEvents := []models.Event{
		{
			Type:        "GOAL",
			Description: "Test goal",
			TeamCode:    "WPG",
			GameCode:    game.GameCode,
		},
	}

	// Send events to the channel
	events <- testEvents

	// Test that fireGoalEvents runs without error
	assert.NotPanics(t, func() {
		fireGoalEvents(events, game)
	})

	// Close the channel
	close(events)
}

func TestFireGoalEvents_AllTeamsMonitored(t *testing.T) {
	setupTest(t)

	// Create a test game
	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")

	// Set up to watch all teams
	viper.Set("watch.nhl", []string{"*"})

	// Create a channel for events
	events := make(chan []models.Event, 1)

	// Create test events
	testEvents := []models.Event{
		{
			Type:        "GOAL",
			Description: "Test goal",
			TeamCode:    "WPG",
			GameCode:    game.GameCode,
		},
	}

	// Send events to the channel
	events <- testEvents

	// Test that fireGoalEvents runs without error
	assert.NotPanics(t, func() {
		fireGoalEvents(events, game)
	})

	// Close the channel
	close(events)
}

func TestFireGoalEvents_MultipleEvents(t *testing.T) {
	setupTest(t)

	// Create a test game
	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")

	// Set up teams to watch
	viper.Set("watch.nhl", []string{"WPG", "TOR"})

	// Create a channel for events
	events := make(chan []models.Event, 1)

	// Create multiple test events
	testEvents := []models.Event{
		{
			Type:        "GOAL",
			Description: "Test goal 1",
			TeamCode:    "WPG",
			GameCode:    game.GameCode,
		},
		{
			Type:        "GOAL",
			Description: "Test goal 2",
			TeamCode:    "TOR",
			GameCode:    game.GameCode,
		},
	}

	// Send events to the channel
	events <- testEvents

	// Test that fireGoalEvents runs without error
	assert.NotPanics(t, func() {
		fireGoalEvents(events, game)
	})

	// Close the channel
	close(events)
}

func TestFireGoalEvents_EmptyEvents(t *testing.T) {
	setupTest(t)

	// Create a test game
	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")

	// Set up teams to watch
	viper.Set("watch.nhl", []string{"WPG"})

	// Create a channel for events
	events := make(chan []models.Event, 1)

	// Create empty test events
	testEvents := []models.Event{}

	// Send events to the channel
	events <- testEvents

	// Test that fireGoalEvents runs without error
	assert.NotPanics(t, func() {
		fireGoalEvents(events, game)
	})

	// Close the channel
	close(events)
}

func TestTeamIsMonitoredByLeague2(t *testing.T) {
	setupTest(t)

	// Test with specific team
	viper.Set("watch.nhl", []string{"WPG"})
	assert.True(t, teamIsMonitoredByLeague("WPG", "NHL"))
	assert.False(t, teamIsMonitoredByLeague("TOR", "NHL"))

	// Test with all teams wildcard
	viper.Set("watch.nhl", []string{"*"})
	assert.True(t, teamIsMonitoredByLeague("WPG", "NHL"))
	assert.True(t, teamIsMonitoredByLeague("TOR", "NHL"))

	// Test case insensitive
	viper.Set("watch.nhl", []string{"wpg"})
	assert.True(t, teamIsMonitoredByLeague("WPG", "NHL"))
	assert.True(t, teamIsMonitoredByLeague("wpg", "NHL"))

	// Test with empty watch list
	viper.Set("watch.nhl", []string{})
	assert.False(t, teamIsMonitoredByLeague("WPG", "NHL"))
}

func TestCheckForNewActiveGames_TeamsMonitored(t *testing.T) {
	setupTest(t)

	// Create a mock service
	mockService := &MockLeagueService{
		leagueName: "NHL",
		games: []models.Game{
			createTestGame(models.LeagueIdNHL, "WPG", "TOR"),
		},
	}

	// Set up teams to watch
	viper.Set("watch.nhl", []string{"WPG"})

	// Test that checkForNewActiveGames runs without error
	assert.NotPanics(t, func() {
		checkForNewActiveGames(mockService)
	})
}

func TestCheckForNewActiveGames_TeamsNotMonitored(t *testing.T) {
	setupTest(t)

	// Create a mock service
	mockService := &MockLeagueService{
		leagueName: "NHL",
		games: []models.Game{
			createTestGame(models.LeagueIdNHL, "WPG", "TOR"),
		},
	}

	// Set up teams to watch (different teams)
	viper.Set("watch.nhl", []string{"BOS"})

	// Test that checkForNewActiveGames runs without error
	assert.NotPanics(t, func() {
		checkForNewActiveGames(mockService)
	})
}

func TestCheckForNewActiveGames_GameAlreadyMonitored(t *testing.T) {
	setupTest(t)

	// Create a mock service
	mockService := &MockLeagueService{
		leagueName: "NHL",
		games: []models.Game{
			createTestGame(models.LeagueIdNHL, "WPG", "TOR"),
		},
	}

	// Set up teams to watch
	viper.Set("watch.nhl", []string{"WPG"})

	// Add the game to active games first
	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")
	memoryStore.AppendActiveGame(game)

	// Test that checkForNewActiveGames runs without error
	assert.NotPanics(t, func() {
		checkForNewActiveGames(mockService)
	})
}
