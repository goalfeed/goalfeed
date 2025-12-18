package nhl

import (
	nhlClients "goalfeed/clients/leagues/nhl"
	"goalfeed/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestMockClientWithCustomScoreboard is a test helper that allows custom scoreboard responses
type TestMockClientWithCustomScoreboard struct {
	nhlClients.MockNHLApiClient
	customScoreboard nhlClients.NHLScoreboardResponse
	useCustom        bool
}

func (m *TestMockClientWithCustomScoreboard) GetNHLScoreBoard(gameId string) nhlClients.NHLScoreboardResponse {
	if m.useCustom {
		return m.customScoreboard
	}
	return m.MockNHLApiClient.GetNHLScoreBoard(gameId)
}

func TestGameFromScoreboard_Overtime(t *testing.T) {
	// Test gameFromScoreboard with overtime period type
	mockClient := &TestMockClientWithCustomScoreboard{
		MockNHLApiClient: nhlClients.MockNHLApiClient{},
		useCustom:        true,
		customScoreboard: nhlClients.NHLScoreboardResponse{
			ID:        2023020193,
			Season:    20232024,
			GameType:  2,
			GameState: "LIVE",
			PeriodDescriptor: nhlClients.PeriodDescriptor{
				Number:     4,
				PeriodType: "OT",
			},
			Clock: nhlClients.Clock{
				TimeRemaining: "05:00",
			},
			HomeTeam: nhlClients.NHLScheduleTeam{
				PlaceName: nhlClients.PlaceName{Default: "Boston"},
				Abbrev:    "BOS",
				Score:     3,
				Sog:       25,
			},
			AwayTeam: nhlClients.NHLScheduleTeam{
				PlaceName: nhlClients.PlaceName{Default: "New York"},
				Abbrev:    "NYI",
				Score:     2,
				Sog:       20,
			},
			Venue: nhlClients.Venue{Default: "TD Garden"},
			StartTimeUTC: time.Now(),
		},
	}
	service := NHLService{Client: mockClient}
	game := service.gameFromScoreboard("2023020193")
	
	assert.Equal(t, 4, game.CurrentState.Period)
	assert.Equal(t, "OVERTIME", game.CurrentState.PeriodType)
	assert.Equal(t, "05:00", game.CurrentState.Clock)
}

func TestGameFromScoreboard_Shootout(t *testing.T) {
	// Test gameFromScoreboard with shootout period type
	mockClient := &TestMockClientWithCustomScoreboard{
		MockNHLApiClient: nhlClients.MockNHLApiClient{},
		useCustom:        true,
		customScoreboard: nhlClients.NHLScoreboardResponse{
			ID:        2023020193,
			Season:    20232024,
			GameType:  2,
			GameState: "LIVE",
			PeriodDescriptor: nhlClients.PeriodDescriptor{
				Number:     5,
				PeriodType: "SO",
			},
			Clock: nhlClients.Clock{
				TimeRemaining: "",
			},
			HomeTeam: nhlClients.NHLScheduleTeam{
				PlaceName: nhlClients.PlaceName{Default: "Boston"},
				Abbrev:    "BOS",
				Score:     3,
				Sog:       25,
			},
			AwayTeam: nhlClients.NHLScheduleTeam{
				PlaceName: nhlClients.PlaceName{Default: "New York"},
				Abbrev:    "NYI",
				Score:     2,
				Sog:       20,
			},
			Venue: nhlClients.Venue{Default: "TD Garden"},
			StartTimeUTC: time.Now(),
		},
	}
	service := NHLService{Client: mockClient}
	game := service.gameFromScoreboard("2023020193")
	
	assert.Equal(t, 5, game.CurrentState.Period)
	assert.Equal(t, "SHOOTOUT", game.CurrentState.PeriodType)
	assert.Equal(t, "LIVE", game.CurrentState.Clock) // Should fallback to LIVE when clock is empty
}

func TestGameFromScoreboard_NoPeriodDescriptor(t *testing.T) {
	// Test gameFromScoreboard when PeriodDescriptor.Number is 0 but game is LIVE
	mockClient := &TestMockClientWithCustomScoreboard{
		MockNHLApiClient: nhlClients.MockNHLApiClient{},
		useCustom:        true,
		customScoreboard: nhlClients.NHLScoreboardResponse{
			ID:        2023020193,
			Season:    20232024,
			GameType:  2,
			GameState: "LIVE",
			PeriodDescriptor: nhlClients.PeriodDescriptor{
				Number:     0, // Missing period descriptor
				PeriodType: "",
			},
			Clock: nhlClients.Clock{
				TimeRemaining: "",
			},
			HomeTeam: nhlClients.NHLScheduleTeam{
				PlaceName: nhlClients.PlaceName{Default: "Boston"},
				Abbrev:    "BOS",
				Score:     3,
				Sog:       25,
			},
			AwayTeam: nhlClients.NHLScheduleTeam{
				PlaceName: nhlClients.PlaceName{Default: "New York"},
				Abbrev:    "NYI",
				Score:     2,
				Sog:       20,
			},
			Venue: nhlClients.Venue{Default: "TD Garden"},
			StartTimeUTC: time.Now(),
		},
	}
	service := NHLService{Client: mockClient}
	game := service.gameFromScoreboard("2023020193")
	
	assert.Equal(t, 1, game.CurrentState.Period) // Should fallback to period 1
	assert.Equal(t, "REGULAR", game.CurrentState.PeriodType)
	assert.Equal(t, "LIVE", game.CurrentState.Clock)
}

func TestGameFromSchedule_DifferentPeriodTypes(t *testing.T) {
	// Test gameFromSchedule with different period types
	mockClient := nhlClients.MockNHLApiClient{}
	service := NHLService{Client: mockClient}
	
	// Get a game from schedule to modify
	schedule := service.getSchedule()
	if len(schedule.GameWeek) == 0 || len(schedule.GameWeek[0].Games) == 0 {
		t.Skip("No games in schedule")
	}
	
	// Test with OT period type
	game := schedule.GameWeek[0].Games[0]
	game.PeriodDescriptor.Number = 4
	game.PeriodDescriptor.PeriodType = "OT"
	result := service.gameFromSchedule(game)
	assert.Equal(t, 4, result.CurrentState.Period)
	assert.Equal(t, "OVERTIME", result.CurrentState.PeriodType)
	
	// Test with SO period type
	game.PeriodDescriptor.PeriodType = "SO"
	result = service.gameFromSchedule(game)
	assert.Equal(t, "SHOOTOUT", result.CurrentState.PeriodType)
	
	// Test with default/unknown period type
	game.PeriodDescriptor.PeriodType = "UNKNOWN"
	result = service.gameFromSchedule(game)
	assert.Equal(t, "REGULAR", result.CurrentState.PeriodType)
}

func TestGameFromSchedule_UpcomingGameStates(t *testing.T) {
	// Test gameFromSchedule with PRE, FUT, OFF game states
	mockClient := nhlClients.MockNHLApiClient{}
	service := NHLService{Client: mockClient}
	
	schedule := service.getSchedule()
	if len(schedule.GameWeek) == 0 || len(schedule.GameWeek[0].Games) == 0 {
		t.Skip("No games in schedule")
	}
	
	game := schedule.GameWeek[0].Games[0]
	
	// Test PRE state
	game.GameState = "PRE"
	game.PeriodDescriptor.Number = 0
	result := service.gameFromSchedule(game)
	assert.Equal(t, models.GameStatus(models.StatusUpcoming), result.CurrentState.Status)
	
	// Test FUT state
	game.GameState = "FUT"
	result = service.gameFromSchedule(game)
	assert.Equal(t, models.GameStatus(models.StatusUpcoming), result.CurrentState.Status)
	
	// Test OFF state
	game.GameState = "OFF"
	result = service.gameFromSchedule(game)
	assert.Equal(t, models.GameStatus(models.StatusUpcoming), result.CurrentState.Status)
}

func TestGameFromSchedule_LiveWithNoPeriod(t *testing.T) {
	// Test gameFromSchedule when LIVE but no period descriptor
	mockClient := nhlClients.MockNHLApiClient{}
	service := NHLService{Client: mockClient}
	
	schedule := service.getSchedule()
	if len(schedule.GameWeek) == 0 || len(schedule.GameWeek[0].Games) == 0 {
		t.Skip("No games in schedule")
	}
	
	game := schedule.GameWeek[0].Games[0]
	game.GameState = "LIVE"
	game.PeriodDescriptor.Number = 0
	game.PeriodDescriptor.PeriodType = ""
	
	result := service.gameFromSchedule(game)
	assert.Equal(t, 1, result.CurrentState.Period) // Should fallback to period 1
	assert.Equal(t, "REGULAR", result.CurrentState.PeriodType)
	assert.Equal(t, "LIVE", result.CurrentState.Clock)
}

func TestGetActiveGames_ScoreboardFallback(t *testing.T) {
	// Test GetActiveGames when scoreboard GameCode doesn't match (fallback path)
	mockClient := &TestMockClientWithCustomScoreboard{
		MockNHLApiClient: nhlClients.MockNHLApiClient{},
		useCustom:        true,
		customScoreboard: nhlClients.NHLScoreboardResponse{
			ID:        999999, // Different ID to trigger fallback
			Season:    20232024,
			GameType:  2,
			GameState: "LIVE",
			PeriodDescriptor: nhlClients.PeriodDescriptor{
				Number:     2,
				PeriodType: "REG",
			},
			Clock: nhlClients.Clock{
				TimeRemaining: "10:00",
			},
			HomeTeam: nhlClients.NHLScheduleTeam{
				PlaceName: nhlClients.PlaceName{Default: "Boston"},
				Abbrev:    "BOS",
				Score:     3,
				Sog:       25,
			},
			AwayTeam: nhlClients.NHLScheduleTeam{
				PlaceName: nhlClients.PlaceName{Default: "New York"},
				Abbrev:    "NYI",
				Score:     2,
				Sog:       20,
			},
			Venue: nhlClients.Venue{Default: "TD Garden"},
			StartTimeUTC: time.Now(),
		},
	}
	service := NHLService{Client: mockClient}
	
	// Get schedule to find a LIVE game
	schedule := service.getSchedule()
	var liveGameID int
	for _, date := range schedule.GameWeek {
		for _, game := range date.Games {
			if game.GameState == "LIVE" {
				liveGameID = game.ID
				break
			}
		}
		if liveGameID > 0 {
			break
		}
	}
	
	if liveGameID == 0 {
		t.Skip("No LIVE games in schedule")
	}
	
	// The scoreboard will return ID 999999, but schedule has different ID
	// This should trigger the fallback to gameFromSchedule
	gamesChan := make(chan []models.Game)
	go service.GetActiveGames(gamesChan)
	activeGames := <-gamesChan
	
	// Should still return games (using fallback)
	assert.Greater(t, len(activeGames), 0)
}

func TestGetActiveGames_NonLiveActiveGame(t *testing.T) {
	// Test GetActiveGames with non-LIVE active game (FINAL state)
	mockClient := nhlClients.MockNHLApiClient{}
	service := NHLService{Client: mockClient}
	
	// This should use gameFromSchedule path for non-LIVE games
	gamesChan := make(chan []models.Game)
	go service.GetActiveGames(gamesChan)
	activeGames := <-gamesChan
	
	// Should return active games
	assert.Greater(t, len(activeGames), 0)
}

func TestGetGameUpdateFromScoreboard_DifferentPeriodTypes(t *testing.T) {
	// Test getGameUpdateFromScoreboard with different period types
	mockClient := &TestMockClientWithCustomScoreboard{
		MockNHLApiClient: nhlClients.MockNHLApiClient{},
		useCustom:        true,
		customScoreboard: nhlClients.NHLScoreboardResponse{
			ID:        2023020193,
			Season:    20232024,
			GameType:  2,
			GameState: "LIVE",
			PeriodDescriptor: nhlClients.PeriodDescriptor{
				Number:     3,
				PeriodType: "OT",
			},
			Clock: nhlClients.Clock{
				TimeRemaining: "03:00",
			},
			HomeTeam: nhlClients.NHLScheduleTeam{
				PlaceName: nhlClients.PlaceName{Default: "Boston"},
				Abbrev:    "BOS",
				Score:     4,
				Sog:       30,
			},
			AwayTeam: nhlClients.NHLScheduleTeam{
				PlaceName: nhlClients.PlaceName{Default: "New York"},
				Abbrev:    "NYI",
				Score:     3,
				Sog:       25,
			},
			Venue: nhlClients.Venue{Default: "TD Garden"},
			StartTimeUTC: time.Now(),
		},
	}
	service := NHLService{Client: mockClient}
	
	// Create a test game
	game := models.Game{
		GameCode: "2023020193",
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team:  models.Team{TeamCode: "BOS"},
				Score: 3,
			},
			Away: models.TeamState{
				Team:  models.Team{TeamCode: "NYI"},
				Score: 2,
			},
		},
	}
	
	updateChan := make(chan models.GameUpdate)
	go service.getGameUpdateFromScoreboard(game, updateChan)
	update := <-updateChan
	
	assert.Equal(t, 3, update.NewState.Period)
	assert.Equal(t, "OVERTIME", update.NewState.PeriodType)
	assert.Equal(t, "03:00", update.NewState.Clock)
	assert.Equal(t, 4, update.NewState.Home.Score)
	assert.Equal(t, 3, update.NewState.Away.Score)
}

func TestGetGameUpdateFromScoreboard_NoPeriodDescriptor(t *testing.T) {
	// Test getGameUpdateFromScoreboard when PeriodDescriptor is missing
	mockClient := &TestMockClientWithCustomScoreboard{
		MockNHLApiClient: nhlClients.MockNHLApiClient{},
		useCustom:        true,
		customScoreboard: nhlClients.NHLScoreboardResponse{
			ID:        2023020193,
			Season:    20232024,
			GameType:  2,
			GameState: "LIVE",
			PeriodDescriptor: nhlClients.PeriodDescriptor{
				Number:     0,
				PeriodType: "",
			},
			Clock: nhlClients.Clock{
				TimeRemaining: "",
			},
			HomeTeam: nhlClients.NHLScheduleTeam{
				PlaceName: nhlClients.PlaceName{Default: "Boston"},
				Abbrev:    "BOS",
				Score:     2,
				Sog:       20,
			},
			AwayTeam: nhlClients.NHLScheduleTeam{
				PlaceName: nhlClients.PlaceName{Default: "New York"},
				Abbrev:    "NYI",
				Score:     1,
				Sog:       15,
			},
			Venue: nhlClients.Venue{Default: "TD Garden"},
			StartTimeUTC: time.Now(),
		},
	}
	service := NHLService{Client: mockClient}
	
	game := models.Game{
		GameCode: "2023020193",
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team:  models.Team{TeamCode: "BOS"},
				Score: 1,
			},
			Away: models.TeamState{
				Team:  models.Team{TeamCode: "NYI"},
				Score: 0,
			},
		},
	}
	
	updateChan := make(chan models.GameUpdate)
	go service.getGameUpdateFromScoreboard(game, updateChan)
	update := <-updateChan
	
	assert.Equal(t, 1, update.NewState.Period) // Should fallback to period 1
	assert.Equal(t, "REGULAR", update.NewState.PeriodType)
	assert.Equal(t, "LIVE", update.NewState.Clock)
}

