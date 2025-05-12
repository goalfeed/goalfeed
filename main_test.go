package main

import (
	"goalfeed/models"
	"goalfeed/services/leagues"
	"goalfeed/targets/memoryStore"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock League Service
type MockLeagueService struct {
	mock.Mock
	leagues.ILeagueService
}

func (m *MockLeagueService) GetLeagueName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockLeagueService) GetActiveGames(gamesChan chan []models.Game) {
	args := m.Called(gamesChan)
	if games, ok := args.Get(0).([]models.Game); ok {
		gamesChan <- games
	}
}

func (m *MockLeagueService) GetGameUpdate(game models.Game, updateChan chan models.GameUpdate) {
	args := m.Called(game, updateChan)
	if update, ok := args.Get(0).(models.GameUpdate); ok {
		updateChan <- update
	}
}

func (m *MockLeagueService) GetEvents(update models.GameUpdate, eventsChan chan []models.Event) {
	args := m.Called(update, eventsChan)
	if events, ok := args.Get(0).([]models.Event); ok {
		eventsChan <- events
	}
}

// Mock Home Assistant
type MockHomeAssistant struct {
	mock.Mock
}

func (m *MockHomeAssistant) SendEvent(event models.Event) error {
	args := m.Called(event)
	return args.Error(0)
}

// Test Helpers
func setupTest(t *testing.T) {
	viper.Reset()
	memoryStore.SetActiveGameKeys([]string{})
	// Reset eventSender to default
	eventSender = func(event models.Event) {
		// Do nothing in tests by default
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

// Existing test
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

func TestInitialize(t *testing.T) {
	setupTest(t)

	// Test initialization of league services
	initialize()

	assert.NotNil(t, leagueServices[models.LeagueIdNHL])
	assert.NotNil(t, leagueServices[models.LeagueIdMLB])
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

func TestCheckForNewActiveGames(t *testing.T) {
	setupTest(t)
	mockService := new(MockLeagueService)

	// Setup test data
	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")
	games := []models.Game{game}

	// Configure mock
	mockService.On("GetLeagueName").Return("nhl")
	mockService.On("GetActiveGames", mock.Anything).Run(func(args mock.Arguments) {
		ch := args.Get(0).(chan []models.Game)
		ch <- games
	}).Return(games)

	// Configure viper for team monitoring
	viper.Set("watch.nhl", []string{"WPG"})

	// Test the function
	checkForNewActiveGames(mockService)

	// Verify the game was added to active games
	assert.True(t, gameIsMonitored(game))
	mockService.AssertExpectations(t)
}

func TestCheckGame(t *testing.T) {
	setupTest(t)
	mockService := new(MockLeagueService)

	// Setup test data
	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")
	memoryStore.SetGame(game)
	memoryStore.AppendActiveGame(game)

	update := models.GameUpdate{
		NewState: models.GameState{
			Status: models.GameStatus(models.StatusEnded),
		},
	}

	events := []models.Event{}

	// Configure mock
	mockService.On("GetLeagueName").Return("nhl")
	mockService.On("GetGameUpdate", mock.AnythingOfType("models.Game"), mock.Anything).Run(func(args mock.Arguments) {
		ch := args.Get(1).(chan models.GameUpdate)
		ch <- update
	}).Return(update)
	mockService.On("GetEvents", mock.AnythingOfType("models.GameUpdate"), mock.Anything).Run(func(args mock.Arguments) {
		ch := args.Get(1).(chan []models.Event)
		ch <- events
	}).Return(events)

	// Set up the league service
	leagueServices[int(models.LeagueIdNHL)] = mockService

	// Test the function
	checkGame(game.GetGameKey())

	// Give the goroutine time to execute
	time.Sleep(100 * time.Millisecond)

	// Verify game was removed from active games when ended
	assert.False(t, gameIsMonitored(game))
	mockService.AssertExpectations(t)
}

func TestFireGoalEvents(t *testing.T) {
	setupTest(t)

	// Setup test data
	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")
	events := []models.Event{
		{TeamCode: "WPG", TeamName: "Winnipeg", LeagueId: int(models.LeagueIdNHL)},
	}

	// Configure viper for team monitoring
	viper.Set("watch.nhl", []string{"WPG"})

	// Create events channel and populate it
	eventsChan := make(chan []models.Event, 1)
	eventsChan <- events

	// Setup mock league service
	mockService := new(MockLeagueService)
	mockService.On("GetLeagueName").Return("nhl")
	leagueServices[int(models.LeagueIdNHL)] = mockService

	// Setup mock event sender
	eventCalled := false
	eventSender = func(event models.Event) {
		eventCalled = true
		assert.Equal(t, "WPG", event.TeamCode)
	}

	// Test the function
	fireGoalEvents(eventsChan, game)

	// Give the goroutine time to execute
	time.Sleep(100 * time.Millisecond)

	// Verify the mock was called
	assert.True(t, eventCalled)
	mockService.AssertExpectations(t)
}

func TestSendTestGoal(t *testing.T) {
	setupTest(t)

	eventCalled := false
	eventSender = func(event models.Event) {
		eventCalled = true
		assert.Equal(t, "TEST", event.TeamCode)
	}

	// Test when test goals are disabled
	viper.Set("test-goals", false)
	sendTestGoal()
	assert.False(t, eventCalled)

	// Test when test goals are enabled
	viper.Set("test-goals", true)
	sendTestGoal()
	time.Sleep(100 * time.Millisecond) // Give the goroutine time to execute
	assert.True(t, eventCalled)
}

func TestRunTickers(t *testing.T) {
	setupTest(t)

	// Create a channel to signal when we want to stop the test
	done := make(chan bool)

	// Start the tickers in a goroutine
	go func() {
		// Let it run for a short time
		time.Sleep(2 * time.Second)
		done <- true
	}()

	// Start the tickers
	go runTickers()

	// Wait for the done signal or timeout
	select {
	case <-done:
		// Test passed
	case <-time.After(3 * time.Second):
		t.Fatal("Test timed out")
	}
}
