package cfl

import (
	"encoding/json"
	"os"
	"testing"

	cflClients "goalfeed/clients/leagues/cfl"
	"goalfeed/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cflPoll(svc CFLService, game models.Game) models.GameUpdate {
	ret := make(chan models.GameUpdate, 1)
	svc.GetGameUpdate(game, ret)
	return <-ret
}

// A failed CFL fetch returns an empty response. Taking it as the new state
// would drop both scores to 0, and recovery would fire one event per point
// already scored.
func TestGetGameUpdate_FailedFetchKeepsPreviousState(t *testing.T) {
	b, err := os.ReadFile("../../../clients/leagues/cfl/testdata/cfl_live_game_numericid.json")
	require.NoError(t, err)
	var live cflClients.CFLLiveGameResponse
	require.NoError(t, json.Unmarshal(b, &live))

	game := models.Game{GameCode: "13419716", LeagueId: models.LeagueIdCFL}
	game.CurrentState = cflPoll(CFLService{Client: cflClients.MockCFLApiClient{LiveGameResponse: live}}, game).NewState
	require.Equal(t, 46, game.CurrentState.Home.Score)

	blip := cflPoll(CFLService{Client: cflClients.MockCFLApiClient{}}, game)
	assert.Equal(t, 46, blip.NewState.Home.Score)
	assert.Equal(t, 16, blip.NewState.Away.Score)
	game.CurrentState = blip.NewState

	recovered := cflPoll(CFLService{Client: cflClients.MockCFLApiClient{LiveGameResponse: live}}, game)
	events := make(chan []models.Event, 1)
	CFLService{}.GetEvents(recovered, events)
	assert.Empty(t, <-events, "recovering from a failed fetch must not fire scores")
}

// Only a response with neither field is a failed fetch. One that is missing
// just one of them is still a real update and must be used: the guard has
// to be "both empty", not "either empty".
func TestGetGameUpdate_OneIdentityFieldIsStillAnUpdate(t *testing.T) {
	b, err := os.ReadFile("../../../clients/leagues/cfl/testdata/cfl_live_game_numericid.json")
	require.NoError(t, err)
	var live cflClients.CFLLiveGameResponse
	require.NoError(t, json.Unmarshal(b, &live))
	require.NotEmpty(t, live.Data.BetGeniusFixtureID)
	require.NotEmpty(t, live.Data.ScoreboardInfo.MatchStatus)

	noStatus := live
	noStatus.Data.ScoreboardInfo.MatchStatus = ""
	noFixture := live
	noFixture.Data.BetGeniusFixtureID = ""
	for name, resp := range map[string]cflClients.CFLLiveGameResponse{"no match status": noStatus, "no fixture id": noFixture} {
		t.Run(name, func(t *testing.T) {
			game := models.Game{GameCode: "13419716", LeagueId: models.LeagueIdCFL}
			got := cflPoll(CFLService{Client: cflClients.MockCFLApiClient{LiveGameResponse: resp}}, game).NewState
			assert.Equal(t, 46, got.Home.Score)
			assert.Equal(t, 16, got.Away.Score)
		})
	}
}
