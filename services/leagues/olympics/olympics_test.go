package olympics

import (
	olympicsClient "goalfeed/clients/leagues/olympics"
	"goalfeed/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOlympicMensHockeyService_GetLeagueName(t *testing.T) {
	client := &olympicsClient.OlympicsApiClient{}
	svc := NewMensHockeyService(client)
	assert.Equal(t, "Olympic Men's Hockey", svc.GetLeagueName())
}

func TestOlympicWomensHockeyService_GetLeagueName(t *testing.T) {
	client := &olympicsClient.OlympicsApiClient{}
	svc := NewWomensHockeyService(client)
	assert.Equal(t, "Olympic Women's Hockey", svc.GetLeagueName())
}

func TestGameCodeRoundTrip(t *testing.T) {
	phase, unit := "GPB-", "000500--"
	gc := gameCode(phase, unit)
	assert.Equal(t, "GPB-_000500--", gc)
	p, u := parseGameCode(gc)
	assert.Equal(t, phase, p)
	assert.Equal(t, unit, u)
}

func TestGetEvents_GoalDiff(t *testing.T) {
	mock := &olympicsClient.MockOlympicsApiClient{}
	svc := NewMensHockeyService(mock)
	svc.Client = mock

	oldState := models.GameState{
		Home: models.TeamState{Team: svc.teamFromCode("CAN", "Canada"), Score: 1},
		Away: models.TeamState{Team: svc.teamFromCode("USA", "United States"), Score: 2},
	}
	newState := models.GameState{
		Home: models.TeamState{Team: svc.teamFromCode("CAN", "Canada"), Score: 3},
		Away: models.TeamState{Team: svc.teamFromCode("USA", "United States"), Score: 2},
	}
	update := models.GameUpdate{OldState: oldState, NewState: newState}

	eventChan := make(chan []models.Event)
	go svc.GetEvents(update, eventChan)
	events := <-eventChan

	// Canada scored 2 more goals
	assert.Len(t, events, 2)
	assert.Equal(t, "CAN", events[0].TeamCode)
	assert.Equal(t, "Canada", events[0].TeamName)
	assert.Equal(t, int(svc.LeagueId), events[0].LeagueId)
	assert.Equal(t, svc.GetLeagueName(), events[0].LeagueName)
	assert.Equal(t, "USA", events[0].OpponentCode)
	assert.Equal(t, models.EventTypeGoal, events[0].Type)
}

func TestGetGamesByDate_ReturnsEmpty(t *testing.T) {
	mock := &olympicsClient.MockOlympicsApiClient{}
	svc := NewMensHockeyService(mock)
	ret := make(chan []models.Game)
	go svc.GetGamesByDate("2026-02-14", ret)
	games := <-ret
	assert.Empty(t, games)
}
