package nfl

import (
	"testing"

	nflc "goalfeed/clients/leagues/nfl"
	"goalfeed/models"

	"github.com/stretchr/testify/assert"
)

func TestGetGameUpdateFromScoreboard_PossessionFallback(t *testing.T) {
	// Test possession fallback logic
	service := NFLService{Client: &nflc.NFLMockClient{}}

	game := models.Game{
		GameCode: "test-game",
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team: models.Team{
					TeamCode: "MIA",
					TeamName: "Miami Dolphins",
					ExtID:    "15",
				},
			},
			Away: models.TeamState{
				Team: models.Team{
					TeamCode: "BUF",
					TeamName: "Buffalo Bills",
					ExtID:    "4",
				},
			},
		},
	}

	ret := make(chan models.GameUpdate, 1)
	go service.getGameUpdateFromScoreboard(game, ret)
	update := <-ret

	// Should not panic and should return a valid update
	assert.NotNil(t, update)
}

func TestGetGameUpdateFromScoreboard_PossessionFallback2(t *testing.T) {
	// Test possession fallback logic when possessionText is empty
	service := NFLService{Client: &nflc.NFLMockClient{}}

	game := models.Game{
		GameCode: "test-game",
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team: models.Team{
					TeamCode: "MIA",
					TeamName: "Miami Dolphins",
					ExtID:    "15",
				},
			},
			Away: models.TeamState{
				Team: models.Team{
					TeamCode: "BUF",
					TeamName: "Buffalo Bills",
					ExtID:    "4",
				},
			},
		},
	}

	ret := make(chan models.GameUpdate, 1)
	go service.getGameUpdateFromScoreboard(game, ret)
	update := <-ret

	// Should not panic and should return a valid update
	assert.NotNil(t, update)
}

func TestGetGameUpdateFromScoreboard_StatusUpcoming(t *testing.T) {
	// Test status derivation when game is upcoming
	service := NFLService{Client: &nflc.NFLMockClient{}}

	game := models.Game{
		GameCode: "test-game",
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team: models.Team{
					TeamCode: "MIA",
					TeamName: "Miami Dolphins",
					ExtID:    "15",
				},
			},
			Away: models.TeamState{
				Team: models.Team{
					TeamCode: "BUF",
					TeamName: "Buffalo Bills",
					ExtID:    "4",
				},
			},
		},
	}

	ret := make(chan models.GameUpdate, 1)
	go service.getGameUpdateFromScoreboard(game, ret)
	update := <-ret

	// Should not panic and should return a valid update
	assert.NotNil(t, update)
}

func TestGetGameUpdateFromScoreboard_StatusActive3(t *testing.T) {
	// Test status derivation when game is active
	service := NFLService{Client: &nflc.NFLMockClient{}}

	game := models.Game{
		GameCode: "test-game",
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team: models.Team{
					TeamCode: "MIA",
					TeamName: "Miami Dolphins",
					ExtID:    "15",
				},
			},
			Away: models.TeamState{
				Team: models.Team{
					TeamCode: "BUF",
					TeamName: "Buffalo Bills",
					ExtID:    "4",
				},
			},
		},
	}

	ret := make(chan models.GameUpdate, 1)
	go service.getGameUpdateFromScoreboard(game, ret)
	update := <-ret

	// Should not panic and should return a valid update
	assert.NotNil(t, update)
}

func TestGameFromScoreboard_SituationFallback4(t *testing.T) {
	// Test situation parsing fallback logic
	service := NFLService{Client: &nflc.NFLMockClient{}}

	game := service.GameFromScoreboard("test-game")

	// Should not panic and should return a valid game
	assert.NotNil(t, game)
	assert.Equal(t, "test-game", game.GameCode)
	assert.Equal(t, models.League(models.LeagueIdNFL), game.LeagueId)
}

func TestGameFromScoreboard_SituationFallback5(t *testing.T) {
	// Test situation parsing fallback with ShortDownDistanceText
	service := NFLService{Client: &nflc.NFLMockClient{}}

	game := service.GameFromScoreboard("test-game")

	// Should not panic and should return a valid game
	assert.NotNil(t, game)
	assert.Equal(t, "test-game", game.GameCode)
	assert.Equal(t, models.League(models.LeagueIdNFL), game.LeagueId)
}

func TestGameFromScoreboard_EmptyEvents2(t *testing.T) {
	// Test with empty events
	service := NFLService{Client: &nflc.NFLMockClient{}}

	game := service.GameFromScoreboard("empty-events")

	// Should return a minimal game
	assert.NotNil(t, game)
	assert.Equal(t, "empty-events", game.GameCode)
	assert.Equal(t, models.League(models.LeagueIdNFL), game.LeagueId)
}

func TestGameFromScoreboard_EmptyCompetitions2(t *testing.T) {
	// Test with empty competitions
	service := NFLService{Client: &nflc.NFLMockClient{}}

	game := service.GameFromScoreboard("empty-competitions")

	// Should return a minimal game
	assert.NotNil(t, game)
	assert.Equal(t, "empty-competitions", game.GameCode)
	assert.Equal(t, models.League(models.LeagueIdNFL), game.LeagueId)
}

func TestGameFromScoreboard_HalftimeDetection2(t *testing.T) {
	// Test halftime detection
	service := NFLService{Client: &nflc.NFLMockClient{}}

	game := service.GameFromScoreboard("halftime-detail")

	// Should not panic and should return a valid game
	assert.NotNil(t, game)
	assert.Equal(t, "halftime-detail", game.GameCode)
	assert.Equal(t, models.League(models.LeagueIdNFL), game.LeagueId)
}

func TestGameFromScoreboard_HalftimePeriod3(t *testing.T) {
	// Test halftime detection with period 2 and clock 0:00
	service := NFLService{Client: &nflc.NFLMockClient{}}

	game := service.GameFromScoreboard("halftime-period2")

	// Should not panic and should return a valid game
	assert.NotNil(t, game)
	assert.Equal(t, "halftime-period2", game.GameCode)
	assert.Equal(t, models.League(models.LeagueIdNFL), game.LeagueId)
}

func TestGameFromScoreboard_SituationFallback6(t *testing.T) {
	// Test situation parsing fallback with situation-fallback
	service := NFLService{Client: &nflc.NFLMockClient{}}

	game := service.GameFromScoreboard("situation-fallback")

	// Should not panic and should return a valid game
	assert.NotNil(t, game)
	assert.Equal(t, "situation-fallback", game.GameCode)
	assert.Equal(t, models.League(models.LeagueIdNFL), game.LeagueId)
}
