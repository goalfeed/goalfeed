package nhl

import (
	nhlClients "goalfeed/clients/leagues/nhl"
	"goalfeed/models"
	"goalfeed/services/leagues"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
}

func TestGetEvents(t *testing.T) {
	var mockClient = nhlClients.MockNHLApiClient{}
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

func getMockService(mockClient nhlClients.MockNHLApiClient) leagues.ILeagueService {
	return NHLService{
		Client: mockClient,
	}
}

func TestGetActiveGames(t *testing.T) {
	var gamesChan chan []models.Game = make(chan []models.Game)
	var mockClient = nhlClients.MockNHLApiClient{}
	service := getMockService(mockClient)
	go service.GetActiveGames(gamesChan)
	activeGames := <-gamesChan
	assert.Equal(t, 9, len(activeGames))
}

func getActiveGame(service leagues.ILeagueService) models.Game {
	var gamesChan chan []models.Game = make(chan []models.Game)
	go service.GetActiveGames(gamesChan)
	activeGames := <-gamesChan
	return activeGames[0]
}
func TestGetGameUpdate(t *testing.T) {
	var updateChan chan models.GameUpdate = make(chan models.GameUpdate)
	var mockClient = nhlClients.MockNHLApiClient{}
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

func TestGetLeagueName(t *testing.T) {
	var mockClient = nhlClients.MockNHLApiClient{}
	service := getMockService(mockClient)
	assert.Equal(t, "NHL", service.GetLeagueName())
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
	// Test ended status codes
	assert.Equal(t, models.GameStatus(models.StatusEnded), gameStatusFromStatusCode("6"))
	assert.Equal(t, models.GameStatus(models.StatusEnded), gameStatusFromStatusCode("7"))
	
	// Test active status codes
	assert.Equal(t, models.GameStatus(models.StatusActive), gameStatusFromStatusCode("1"))
	assert.Equal(t, models.GameStatus(models.StatusActive), gameStatusFromStatusCode("2"))
	assert.Equal(t, models.GameStatus(models.StatusActive), gameStatusFromStatusCode("3"))
	assert.Equal(t, models.GameStatus(models.StatusActive), gameStatusFromStatusCode("unknown"))
}

func TestGameStatusFromGameState(t *testing.T) {
	// Test all different game states
	assert.Equal(t, models.GameStatus(models.StatusUpcoming), gameStatusFromGameState("PRE"))
	assert.Equal(t, models.GameStatus(models.StatusUpcoming), gameStatusFromGameState("OFF"))
	assert.Equal(t, models.GameStatus(models.StatusUpcoming), gameStatusFromGameState("FUT"))
	assert.Equal(t, models.GameStatus(models.StatusActive), gameStatusFromGameState("LIVE"))
	assert.Equal(t, models.GameStatus(models.StatusEnded), gameStatusFromGameState("FINAL"))
	assert.Equal(t, models.GameStatus(models.StatusActive), gameStatusFromGameState("Unknown State"))
}
