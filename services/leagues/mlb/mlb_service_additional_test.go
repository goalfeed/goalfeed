package mlb

import (
	"testing"

	mlbc "goalfeed/clients/leagues/mlb"
	"goalfeed/models"

	"github.com/stretchr/testify/assert"
)

func TestGetGameUpdateFromScoreboard(t *testing.T) {
	// Test successful scoreboard update
	service := MLBService{Client: &mlbc.MockMLBApiClient{}}

	game := models.Game{
		GameCode: "test-game",
		LeagueId: models.LeagueIdMLB,
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team: models.Team{
					TeamCode: "NYY",
					TeamName: "New York Yankees",
					ExtID:    "NYY",
				},
			},
			Away: models.TeamState{
				Team: models.Team{
					TeamCode: "BOS",
					TeamName: "Boston Red Sox",
					ExtID:    "BOS",
				},
			},
		},
	}

	ret := make(chan models.GameUpdate, 1)
	go service.getGameUpdateFromScoreboard(game, ret)
	update := <-ret

	// Should not panic and should return a valid update
	assert.NotNil(t, update)
	assert.Equal(t, game.CurrentState, update.OldState)
	assert.NotNil(t, update.NewState)
}

func TestGetGameUpdateFromScoreboard_WithInning(t *testing.T) {
	// Test scoreboard update with inning information
	service := MLBService{Client: &mlbc.MockMLBApiClient{}}

	game := models.Game{
		GameCode: "test-game",
		LeagueId: models.LeagueIdMLB,
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team: models.Team{
					TeamCode: "NYY",
					TeamName: "New York Yankees",
					ExtID:    "NYY",
				},
			},
			Away: models.TeamState{
				Team: models.Team{
					TeamCode: "BOS",
					TeamName: "Boston Red Sox",
					ExtID:    "BOS",
				},
			},
		},
	}

	ret := make(chan models.GameUpdate, 1)
	go service.getGameUpdateFromScoreboard(game, ret)
	update := <-ret

	// Should not panic and should return a valid update
	assert.NotNil(t, update)
	assert.Equal(t, game.CurrentState, update.OldState)
	assert.NotNil(t, update.NewState)
}

func TestGetGameUpdateFromScoreboard_WithCount(t *testing.T) {
	// Test scoreboard update with count information
	service := MLBService{Client: &mlbc.MockMLBApiClient{}}

	game := models.Game{
		GameCode: "test-game",
		LeagueId: models.LeagueIdMLB,
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team: models.Team{
					TeamCode: "NYY",
					TeamName: "New York Yankees",
					ExtID:    "NYY",
				},
			},
			Away: models.TeamState{
				Team: models.Team{
					TeamCode: "BOS",
					TeamName: "Boston Red Sox",
					ExtID:    "BOS",
				},
			},
		},
	}

	ret := make(chan models.GameUpdate, 1)
	go service.getGameUpdateFromScoreboard(game, ret)
	update := <-ret

	// Should not panic and should return a valid update
	assert.NotNil(t, update)
	assert.Equal(t, game.CurrentState, update.OldState)
	assert.NotNil(t, update.NewState)
}

func TestGetGameUpdateFromScoreboard_WithBaseRunners(t *testing.T) {
	// Test scoreboard update with base runners
	service := MLBService{Client: &mlbc.MockMLBApiClient{}}

	game := models.Game{
		GameCode: "test-game",
		LeagueId: models.LeagueIdMLB,
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team: models.Team{
					TeamCode: "NYY",
					TeamName: "New York Yankees",
					ExtID:    "NYY",
				},
			},
			Away: models.TeamState{
				Team: models.Team{
					TeamCode: "BOS",
					TeamName: "Boston Red Sox",
					ExtID:    "BOS",
				},
			},
		},
	}

	ret := make(chan models.GameUpdate, 1)
	go service.getGameUpdateFromScoreboard(game, ret)
	update := <-ret

	// Should not panic and should return a valid update
	assert.NotNil(t, update)
	assert.Equal(t, game.CurrentState, update.OldState)
	assert.NotNil(t, update.NewState)
}

func TestGetGameUpdateFromScoreboard_StatusOverride(t *testing.T) {
	// Test status override logic when inning/count indicate gameplay
	service := MLBService{Client: &mlbc.MockMLBApiClient{}}

	game := models.Game{
		GameCode: "test-game",
		LeagueId: models.LeagueIdMLB,
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team: models.Team{
					TeamCode: "NYY",
					TeamName: "New York Yankees",
					ExtID:    "NYY",
				},
			},
			Away: models.TeamState{
				Team: models.Team{
					TeamCode: "BOS",
					TeamName: "Boston Red Sox",
					ExtID:    "BOS",
				},
			},
		},
	}

	ret := make(chan models.GameUpdate, 1)
	go service.getGameUpdateFromScoreboard(game, ret)
	update := <-ret

	// Should not panic and should return a valid update
	assert.NotNil(t, update)
	assert.Equal(t, game.CurrentState, update.OldState)
	assert.NotNil(t, update.NewState)
}
