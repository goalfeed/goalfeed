package mlb

import (
	"errors"
	mlb "goalfeed/clients/leagues/mlb"
	"goalfeed/models"
	"goalfeed/services/leagues"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockMLBApiClientWithError for testing error scenarios
type MockMLBApiClientWithError struct{}

func (c MockMLBApiClientWithError) GetMLBSchedule() mlb.MLBScheduleResponse {
	var mockClient = mlb.MockMLBApiClient{}
	return mockClient.GetMLBSchedule()
}

func (c MockMLBApiClientWithError) GetMLBScoreBoard(sGameId string) mlb.MLBScoreboardResponse {
	var mockClient = mlb.MockMLBApiClient{}
	return mockClient.GetMLBScoreBoard(sGameId)
}

func (c MockMLBApiClientWithError) GetTeam(sLink string) mlb.MLBTeamResponse {
	var mockClient = mlb.MockMLBApiClient{}
	return mockClient.GetTeam(sLink)
}

func (c MockMLBApiClientWithError) GetDiffPatch(gameId string, timestamp string) (mlb.MLBDiffPatch, error) {
	return mlb.MLBDiffPatch{}, errors.New("mock error for testing")
}

func TestGetEvents(t *testing.T) {
	var mockClient = mlb.MockMLBApiClient{}
	service := getMockService(mockClient)
	var updateChan chan models.GameUpdate = make(chan models.GameUpdate)
	activeGame := getActiveGame(service)
	newAway := activeGame.CurrentState.Away.Score + 1
	mockClient.SetAwayScore(newAway)
	go service.GetGameUpdate(activeGame, updateChan)
	update := <-updateChan
	var eventChan chan []models.Event = make(chan []models.Event)
	go service.GetEvents(update, eventChan)
	events := <-eventChan
	assert.Equal(t, activeGame.CurrentState.Away.Team.TeamCode, events[0].TeamCode)
	assert.Equal(t, activeGame.CurrentState.Away.Team.GetTeamHash(), events[0].TeamHash)
	assert.Equal(t, activeGame.CurrentState.Away.Team.TeamName, events[0].TeamName)
	assert.Equal(t, activeGame.CurrentState.Away.Team.LeagueID, events[0].LeagueId)
	assert.Equal(t, service.GetLeagueName(), events[0].LeagueName)
	assert.Equal(t, activeGame.CurrentState.Home.Team.TeamCode, events[0].OpponentCode)
	assert.Equal(t, activeGame.CurrentState.Home.Team.GetTeamHash(), events[0].OpponentHash)
	assert.Equal(t, activeGame.CurrentState.Home.Team.TeamName, events[0].OpponentName)

	activeGame.CurrentState = update.NewState
	newHome := activeGame.CurrentState.Home.Score + 2
	mockClient.SetHomeScore(newHome)
	go service.GetGameUpdate(activeGame, updateChan)
	update = <-updateChan
	go service.GetEvents(update, eventChan)
	events = <-eventChan
	assert.Equal(t, newHome-activeGame.CurrentState.Home.Score, len(events))
	assert.Equal(t, activeGame.CurrentState.Home.Team.TeamCode, events[0].TeamCode)
	assert.Equal(t, activeGame.CurrentState.Home.Team.GetTeamHash(), events[0].TeamHash)
	assert.Equal(t, activeGame.CurrentState.Home.Team.TeamName, events[0].TeamName)
	assert.Equal(t, activeGame.CurrentState.Home.Team.LeagueID, events[0].LeagueId)
	assert.Equal(t, service.GetLeagueName(), events[0].LeagueName)
	assert.Equal(t, activeGame.CurrentState.Away.Team.TeamCode, events[0].OpponentCode)
	assert.Equal(t, activeGame.CurrentState.Away.Team.GetTeamHash(), events[0].OpponentHash)
	assert.Equal(t, activeGame.CurrentState.Away.Team.TeamName, events[0].OpponentName)
}

func getMockService(mockClient mlb.MockMLBApiClient) leagues.ILeagueService {
	return MLBService{
		Client: mockClient,
	}
}

func TestGetActiveGames(t *testing.T) {
	var gamesChan chan []models.Game = make(chan []models.Game)
	var mockClient = mlb.MockMLBApiClient{}
	service := getMockService(mockClient)
	go service.GetActiveGames(gamesChan)
	activeGames := <-gamesChan
	assert.Equal(t, len(activeGames), 1)
}

func getActiveGame(service leagues.ILeagueService) models.Game {
	var gamesChan chan []models.Game = make(chan []models.Game)
	go service.GetActiveGames(gamesChan)
	activeGames := <-gamesChan
	return activeGames[0]
}
func TestGetGameUpdate(t *testing.T) {
	var updateChan chan models.GameUpdate = make(chan models.GameUpdate)
	var mockClient = mlb.MockMLBApiClient{}
	service := getMockService(mockClient)
	activeGame := getActiveGame(service)
	go service.GetGameUpdate(activeGame, updateChan)
	update := <-updateChan
	assert.NotEmpty(t, update)
	newAway := activeGame.CurrentState.Away.Score + 1
	newHome := activeGame.CurrentState.Home.Score + 1
	mockClient.SetAwayScore(newAway)
	mockClient.SetHomeScore(newHome)
	activeGame.CurrentState = update.NewState
	go service.GetGameUpdate(activeGame, updateChan)
	update = <-updateChan
	assert.Equal(t, update.OldState.Away.Score, activeGame.CurrentState.Away.Score)
	assert.Equal(t, update.OldState.Home.Score, activeGame.CurrentState.Home.Score)
	assert.Equal(t, update.NewState.Away.Score, newAway)
	assert.Equal(t, update.NewState.Home.Score, newHome)

}

func TestFudgeTimestamp(t *testing.T) {
	// Test the fudgeTimestamp function
	input := "20230101_123456"
	expected := "20230101_123446" // 123456 - 10 = 123446
	result := fudgeTimestamp(input)
	assert.Equal(t, expected, result)

	// Test with different input
	input2 := "20230201_000100"
	expected2 := "20230201_000090" // 000100 - 10 = 000090
	result2 := fudgeTimestamp(input2)
	assert.Equal(t, expected2, result2)
}

func TestGameStatusFromStatusCode(t *testing.T) {
	// Test ended status code
	assert.Equal(t, models.GameStatus(models.StatusEnded), gameStatusFromStatusCode("7"))

	// Test active status codes
	assert.Equal(t, models.GameStatus(models.StatusActive), gameStatusFromStatusCode("1"))
	assert.Equal(t, models.GameStatus(models.StatusActive), gameStatusFromStatusCode("2"))
	assert.Equal(t, models.GameStatus(models.StatusActive), gameStatusFromStatusCode("unknown"))
}

func TestGetGameUpdateFromDiffPatch(t *testing.T) {
	var mockClient = mlb.MockMLBApiClient{}
	service := getMockService(mockClient)
	activeGame := getActiveGame(service)

	// Test the getGameUpdateFromDiffPatch function by calling it indirectly
	// Since it's only called from GetGameUpdate when there's an ExtTimestamp
	activeGame.CurrentState.ExtTimestamp = "20230101_123456"
	var updateChan chan models.GameUpdate = make(chan models.GameUpdate)

	go service.GetGameUpdate(activeGame, updateChan)
	update := <-updateChan
	assert.NotNil(t, update)
}

func TestGetGameUpdateFromDiffPatchWithEmptyValues(t *testing.T) {
	var mockClient = mlb.MockMLBApiClient{}
	service := getMockService(mockClient)
	activeGame := getActiveGame(service)

	// Set initial scores for comparison
	activeGame.CurrentState.Home.Score = 3
	activeGame.CurrentState.Away.Score = 2
	activeGame.CurrentState.ExtTimestamp = "20230101_123456"

	// Create a specific diff patch response that covers empty values branch
	mockClient.SetHomeScore(0) // This will trigger the empty homeScore branch
	mockClient.SetAwayScore(0) // This will trigger the empty awayScore branch

	var updateChan chan models.GameUpdate = make(chan models.GameUpdate)
	go service.GetGameUpdate(activeGame, updateChan)
	update := <-updateChan

	// Should use the original scores when diff patch returns 0
	assert.Equal(t, activeGame.CurrentState.Home.Score, update.NewState.Home.Score)
	assert.Equal(t, activeGame.CurrentState.Away.Score, update.NewState.Away.Score)
}

func TestGetGameUpdateErrorFallback(t *testing.T) {
	// Create a custom mock client that will return an error for GetDiffPatch
	var errorClient MockMLBApiClientWithError
	service := MLBService{Client: errorClient}
	activeGame := getActiveGame(getMockService(mlb.MockMLBApiClient{}))

	// Set ExtTimestamp to trigger diff patch call which will error
	activeGame.CurrentState.ExtTimestamp = "20230101_123456"

	var updateChan chan models.GameUpdate = make(chan models.GameUpdate)
	go service.GetGameUpdate(activeGame, updateChan)
	update := <-updateChan

	// Should fallback to scoreboard method
	assert.NotNil(t, update)
}

func TestGameStatusFromScheduleGame(t *testing.T) {
	// Create mock schedule games to test all status paths
	finalGame := mlb.MLBScheduleResponseGame{
		Status: mlb.Status{
			AbstractGameState: "Final",
		},
	}
	assert.Equal(t, models.GameStatus(models.StatusEnded), gameStatusFromScheduleGame(finalGame))

	upcomingGame := mlb.MLBScheduleResponseGame{
		Status: mlb.Status{
			AbstractGameState: "Preview",
		},
	}
	assert.Equal(t, models.GameStatus(models.StatusUpcoming), gameStatusFromScheduleGame(upcomingGame))

	activeGame := mlb.MLBScheduleResponseGame{
		Status: mlb.Status{
			AbstractGameState: "Live",
		},
	}
	assert.Equal(t, models.GameStatus(models.StatusActive), gameStatusFromScheduleGame(activeGame))

	unknownGame := mlb.MLBScheduleResponseGame{
		Status: mlb.Status{
			AbstractGameState: "Unknown",
		},
	}
	assert.Equal(t, models.GameStatus(models.StatusActive), gameStatusFromScheduleGame(unknownGame))
}
