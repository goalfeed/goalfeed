package iihf

import (
	iihfClients "goalfeed/clients/leagues/iihf"
	"goalfeed/models"
	"goalfeed/services/leagues"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEvents(t *testing.T) {
	var mockClient = iihfClients.MockIIHFApiClient{}
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

func getMockService(mockClient iihfClients.MockIIHFApiClient) leagues.ILeagueService {
	return IIHFService{
		Client: mockClient,
	}
}

func TestGetActiveGames(t *testing.T) {
	var gamesChan chan []models.Game = make(chan []models.Game)
	var mockClient = iihfClients.MockIIHFApiClient{}
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
	var mockClient = iihfClients.MockIIHFApiClient{}
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
