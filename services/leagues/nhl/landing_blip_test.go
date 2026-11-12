package nhl

import (
	"testing"

	nhlClients "goalfeed/clients/leagues/nhl"
	"goalfeed/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A failed landing fetch (network error, or a body that is not JSON) gives
// GetNHLScoreBoard a zero response. That must not become the game's state:
// the next good poll would see the score jump from 0 back to its real value
// and fire that many false goal events.
func TestGetGameUpdate_FailedLandingFetchKeepsPreviousState(t *testing.T) {
	resetRightRailThrottle(t)
	var landing nhlClients.NHLScoreboardResponse
	loadJSON(t, "landing_2026020291_final.json", &landing)
	landing.GameState = "LIVE"
	client := &fixtureClient{landing: landing}
	svc := NHLService{Client: client}

	game := models.Game{GameCode: "2026020291", LeagueId: models.LeagueIdNHL}
	game.CurrentState = poll(t, svc, game).NewState
	require.Greater(t, game.CurrentState.Home.Score+game.CurrentState.Away.Score, 0, "fixture should have goals")
	good := game.CurrentState

	client.landing = nhlClients.NHLScoreboardResponse{} // the fetch failed
	blip := poll(t, svc, game)
	assert.Equal(t, good.Home.Score, blip.NewState.Home.Score)
	assert.Equal(t, good.Away.Score, blip.NewState.Away.Score)
	assert.Equal(t, good.Period, blip.NewState.Period)
	assert.Equal(t, good.Status, blip.NewState.Status)
	game.CurrentState = blip.NewState

	client.landing = landing // recovered
	recovered := poll(t, svc, game)
	events := make(chan []models.Event, 1)
	svc.GetEvents(recovered, events)
	assert.Empty(t, <-events, "recovering from a failed fetch must not fire goals")
}

// A response for some other game is no better than an empty one.
func TestGetGameUpdate_LandingForAnotherGameKeepsPreviousState(t *testing.T) {
	resetRightRailThrottle(t)
	var landing nhlClients.NHLScoreboardResponse
	loadJSON(t, "landing_2026020291_final.json", &landing)
	landing.GameState = "LIVE"
	client := &fixtureClient{landing: landing}
	svc := NHLService{Client: client}

	game := models.Game{GameCode: "2026020291", LeagueId: models.LeagueIdNHL}
	game.CurrentState = poll(t, svc, game).NewState

	other := landing
	other.ID = 2026020001
	other.HomeTeam.Score, other.AwayTeam.Score = 0, 0
	client.landing = other
	got := poll(t, svc, game).NewState
	assert.Equal(t, game.CurrentState.Home.Score, got.Home.Score)
	assert.Equal(t, game.CurrentState.Away.Score, got.Away.Score)
}

func poll(t *testing.T, svc NHLService, game models.Game) models.GameUpdate {
	t.Helper()
	ret := make(chan models.GameUpdate, 1)
	svc.GetGameUpdate(game, ret)
	return <-ret
}
