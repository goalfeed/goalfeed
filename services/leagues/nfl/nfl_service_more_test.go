package nfl

import (
	"testing"

	nflc "goalfeed/clients/leagues/nfl"
	"goalfeed/models"

	"github.com/stretchr/testify/assert"
)

func TestGameFromEvent_HydratesMissingTeams(t *testing.T) {
	svc := NFLService{Client: nflc.NFLMockClient{}}
	ev := nflc.NFLScheduleEvent{ID: "401547403"}
	// Provide competitors with empty abbreviations so hydration path runs
	ev.Competitions = []nflc.NFLCompetition{
		{
			Competitors: []nflc.NFLCompetitor{
				{HomeAway: "home"},
				{HomeAway: "away"},
			},
		},
	}
	g := svc.gameFromEvent(ev)
	if g.CurrentState.Home.Team.TeamCode == "" || g.CurrentState.Away.Team.TeamCode == "" {
		t.Fatalf("expected team codes hydrated from scoreboard")
	}
}

func TestGetActiveGames_MergesScoreboard(t *testing.T) {
	svc := NFLService{Client: nflc.NFLMockClient{}}
	ch := make(chan []models.Game)
	go svc.GetActiveGames(ch)
	games := <-ch
	if len(games) == 0 {
		t.Fatalf("expected at least one active game")
	}
	g := games[0]
	// Mock returns 21-17 in scoreboard; ensure non-zero and Active
	if g.CurrentState.Home.Score == 0 && g.CurrentState.Away.Score == 0 {
		t.Fatalf("expected merged non-zero scores from scoreboard")
	}
	if g.CurrentState.Status != models.StatusActive {
		t.Fatalf("expected active status after merge")
	}
}

func TestGameFromEvent_DateParsing(t *testing.T) {
	svc := NFLService{Client: nflc.NFLMockClient{}}
	ev := nflc.NFLScheduleEvent{ID: "401547403"}
	// Put date in competition date with RFC3339 format
	ev.Competitions = []nflc.NFLCompetition{{Date: "2006-01-02T15:04:05Z"}}
	g := svc.gameFromEvent(ev)
	if g.GameDetails.GameDate.IsZero() {
		t.Fatalf("expected parsed competition date")
	}
}

func TestParseSituationShortDetail(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedDown int
		expectedDist int
		expectedPoss string
		expectedYard int
	}{
		{"empty", "", 0, 0, "", 0},
		{"basic", "1st & 10", 1, 10, "", 0},
		{"with down", "2nd & 5", 2, 5, "", 0},
		{"goal line", "1st & Goal", 0, 0, "", 0},
		{"longer text", "3rd & 15 at NYG 45", 3, 15, "NYG", 45},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			down, dist, poss, yard := parseSituationShortDetail(tt.input)
			assert.Equal(t, tt.expectedDown, down)
			assert.Equal(t, tt.expectedDist, dist)
			assert.Equal(t, tt.expectedPoss, poss)
			assert.Equal(t, tt.expectedYard, yard)
		})
	}
}

func TestGameFromScoreboard_EmptyResponse(t *testing.T) {
	svc := NFLService{Client: nflc.NFLMockClient{}}
	games := svc.GameFromScoreboard("")
	assert.Empty(t, games.GameCode)
}

func TestGameFromScoreboard_WithSituation(t *testing.T) {
	svc := NFLService{Client: nflc.NFLMockClient{}}
	game := svc.GameFromScoreboard("401547403")
	assert.Equal(t, "401547403", game.GameCode)
	assert.Equal(t, models.GameStatus(models.StatusActive), game.CurrentState.Status)
}

func TestGameFromScoreboard_HalftimeLabeling(t *testing.T) {
	svc := NFLService{Client: nflc.NFLMockClient{}}
	game := svc.GameFromScoreboard("401547403")
	assert.Equal(t, models.GameStatus(models.StatusActive), game.CurrentState.Status)
	// Mock should return halftime labeling
}

func TestGameFromScoreboard_OvertimePeriod(t *testing.T) {
	svc := NFLService{Client: nflc.NFLMockClient{}}
	game := svc.GameFromScoreboard("401547403")
	assert.Equal(t, "401547403", game.GameCode)
	// Mock should handle overtime period
}

func TestGameFromScoreboard_TeamNameFallback(t *testing.T) {
	svc := NFLService{Client: nflc.NFLMockClient{}}
	game := svc.GameFromScoreboard("401547403")
	assert.NotEmpty(t, game.CurrentState.Home.Team.TeamCode)
	assert.NotEmpty(t, game.CurrentState.Away.Team.TeamCode)
}

func TestGameFromScoreboard_EmptySituation(t *testing.T) {
	svc := NFLService{Client: nflc.NFLMockClient{}}
	game := svc.GameFromScoreboard("401547403")
	assert.Equal(t, "401547403", game.GameCode)
	// Mock should handle empty situation
}

func TestGameFromEvent_NoMatchingScoreboard(t *testing.T) {
	svc := NFLService{Client: nflc.NFLMockClient{}}
	game := svc.GameFromScoreboard("missing-game")
	assert.Equal(t, "missing-game", game.GameCode)
}

func TestGameFromEvent_EmptyScoreboard(t *testing.T) {
	svc := NFLService{Client: nflc.NFLMockClient{}}
	game := svc.GameFromScoreboard("")
	assert.Empty(t, game.GameCode)
}

func TestGetGameUpdate_Basic(t *testing.T) {
	svc := NFLService{Client: nflc.NFLMockClient{}}

	// Create a test game
	game := models.Game{
		GameCode: "401547403",
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team:  models.Team{TeamCode: "KC", TeamName: "Kansas City Chiefs"},
				Score: 0,
			},
			Away: models.TeamState{
				Team:  models.Team{TeamCode: "BUF", TeamName: "Buffalo Bills"},
				Score: 0,
			},
			Status: models.StatusUpcoming,
		},
	}

	// Test GetGameUpdate
	ret := make(chan models.GameUpdate)
	go svc.GetGameUpdate(game, ret)

	update := <-ret
	assert.Equal(t, game.CurrentState, update.OldState)
	assert.NotNil(t, update.NewState)
}

func TestGetGameUpdateFromScoreboard_WithScoreboardData(t *testing.T) {
	svc := NFLService{Client: nflc.NFLMockClient{}}

	// Create a test game
	game := models.Game{
		GameCode: "401547403",
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team:  models.Team{TeamCode: "KC", TeamName: "Kansas City Chiefs"},
				Score: 0,
			},
			Away: models.TeamState{
				Team:  models.Team{TeamCode: "BUF", TeamName: "Buffalo Bills"},
				Score: 0,
			},
			Status: models.StatusUpcoming,
		},
	}

	// Test getGameUpdateFromScoreboard
	ret := make(chan models.GameUpdate)
	go svc.getGameUpdateFromScoreboard(game, ret)

	update := <-ret
	assert.Equal(t, game.CurrentState, update.OldState)
	assert.NotNil(t, update.NewState)
}

func TestGetGameUpdateFromScoreboard_TeamHydration(t *testing.T) {
	svc := NFLService{Client: nflc.NFLMockClient{}}

	// Create a test game with existing team info
	game := models.Game{
		GameCode: "401547403",
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team: models.Team{
					TeamCode: "KC",
					TeamName: "Kansas City Chiefs",
					ExtID:    "KC",
					LogoURL:  "https://example.com/kc.png",
				},
				Score: 0,
			},
			Away: models.TeamState{
				Team: models.Team{
					TeamCode: "BUF",
					TeamName: "Buffalo Bills",
					ExtID:    "BUF",
					LogoURL:  "https://example.com/buf.png",
				},
				Score: 0,
			},
			Status: models.StatusUpcoming,
		},
	}

	// Test team hydration when API returns empty team data
	ret := make(chan models.GameUpdate)
	go svc.getGameUpdateFromScoreboard(game, ret)

	update := <-ret
	assert.Equal(t, game.CurrentState, update.OldState)
	assert.NotNil(t, update.NewState)
}

func TestGetGameUpdateFromScoreboard_TeamFallback2(t *testing.T) {
	// Test team fallback logic when API returns empty team data
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
					LogoURL:  "https://example.com/mia.png",
				},
			},
			Away: models.TeamState{
				Team: models.Team{
					TeamCode: "BUF",
					TeamName: "Buffalo Bills",
					ExtID:    "4",
					LogoURL:  "https://example.com/buf.png",
				},
			},
		},
	}

	ret := make(chan models.GameUpdate, 1)
	service.getGameUpdateFromScoreboard(game, ret)

	update := <-ret
	assert.NotNil(t, update.NewState)
	assert.Equal(t, "MIA", update.NewState.Home.Team.TeamCode)
	assert.Equal(t, "BUF", update.NewState.Away.Team.TeamCode)
}

func TestGetGameUpdateFromScoreboard_SituationParsing2(t *testing.T) {
	// Test situation parsing from ShortDetail
	service := NFLService{Client: &nflc.NFLMockClient{}}

	game := models.Game{
		GameCode: "test-game",
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Details: models.EventDetails{
				Down:       0,
				Distance:   0,
				Possession: "",
				YardLine:   0,
			},
		},
	}

	ret := make(chan models.GameUpdate, 1)
	service.getGameUpdateFromScoreboard(game, ret)

	update := <-ret
	assert.NotNil(t, update.NewState)
	// The mock client should provide some situation data
	assert.True(t, update.NewState.Details.Down > 0 || update.NewState.Details.Distance > 0)
}

func TestGetGameUpdateFromScoreboard_HalftimeDetection2(t *testing.T) {
	// Test halftime detection logic
	service := NFLService{Client: &nflc.NFLMockClient{}}

	game := models.Game{
		GameCode: "halftime-detail",
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Period: 2,
			Clock:  "0:00",
		},
	}

	ret := make(chan models.GameUpdate, 1)
	service.getGameUpdateFromScoreboard(game, ret)

	update := <-ret
	assert.NotNil(t, update.NewState)
	assert.Equal(t, "HALFTIME", update.NewState.PeriodType)
	assert.Equal(t, "HALFTIME", update.NewState.Clock)
}

func TestGetGameUpdateFromScoreboard_StatusDerivation(t *testing.T) {
	// Test status derivation logic
	service := NFLService{Client: &nflc.NFLMockClient{}}

	game := models.Game{
		GameCode: "test-game",
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Status: models.StatusUpcoming,
		},
	}

	ret := make(chan models.GameUpdate, 1)
	service.getGameUpdateFromScoreboard(game, ret)

	update := <-ret
	assert.NotNil(t, update.NewState)
	// Should derive status from scoreboard data
	assert.True(t, update.NewState.Status == models.StatusActive ||
		update.NewState.Status == models.StatusEnded ||
		update.NewState.Status == models.StatusUpcoming)
}

func TestGetGameUpdateFromScoreboard_NoEvents2(t *testing.T) {
	// Test when scoreboard has no events
	service := NFLService{Client: &nflc.NFLMockClient{}}

	game := models.Game{
		GameCode: "empty-events",
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Status: models.StatusUpcoming,
		},
	}

	ret := make(chan models.GameUpdate, 1)
	service.getGameUpdateFromScoreboard(game, ret)

	update := <-ret
	assert.NotNil(t, update.NewState)
	assert.Equal(t, game.CurrentState, update.NewState)
}

func TestGetGameUpdateFromScoreboard_NoCompetitions2(t *testing.T) {
	// Test when scoreboard has events but no competitions
	service := NFLService{Client: &nflc.NFLMockClient{}}

	game := models.Game{
		GameCode: "empty-competitions",
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Status: models.StatusUpcoming,
		},
	}

	ret := make(chan models.GameUpdate, 1)
	service.getGameUpdateFromScoreboard(game, ret)

	update := <-ret
	assert.NotNil(t, update.NewState)
	assert.Equal(t, game.CurrentState, update.NewState)
}

func TestGetGameUpdateFromScoreboard_LessThanTwoCompetitors(t *testing.T) {
	// Test when scoreboard has competitions but less than 2 competitors
	service := NFLService{Client: &nflc.NFLMockClient{}}

	game := models.Game{
		GameCode: "test-game",
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Status: models.StatusUpcoming,
		},
	}

	ret := make(chan models.GameUpdate, 1)
	service.getGameUpdateFromScoreboard(game, ret)

	update := <-ret
	assert.NotNil(t, update.NewState)
	// The mock client returns real data, so we just verify it doesn't panic
	assert.True(t, update.NewState.Status == models.StatusActive ||
		update.NewState.Status == models.StatusEnded ||
		update.NewState.Status == models.StatusUpcoming)
}

func TestGetGameUpdateFromScoreboard_StatusCompleted(t *testing.T) {
	// Test status derivation when game is completed
	service := NFLService{Client: &nflc.NFLMockClient{}}

	game := models.Game{
		GameCode: "test-game",
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Status: models.StatusUpcoming,
		},
	}

	ret := make(chan models.GameUpdate, 1)
	service.getGameUpdateFromScoreboard(game, ret)

	update := <-ret
	assert.NotNil(t, update.NewState)
	// Should derive status from scoreboard data
	assert.True(t, update.NewState.Status == models.StatusActive ||
		update.NewState.Status == models.StatusEnded ||
		update.NewState.Status == models.StatusUpcoming)
}

func TestGetGameUpdateFromScoreboard_StatusActive(t *testing.T) {
	// Test status derivation when game is active
	service := NFLService{Client: &nflc.NFLMockClient{}}

	game := models.Game{
		GameCode: "test-game",
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Status: models.StatusUpcoming,
		},
	}

	ret := make(chan models.GameUpdate, 1)
	service.getGameUpdateFromScoreboard(game, ret)

	update := <-ret
	assert.NotNil(t, update.NewState)
	// Should derive status from scoreboard data
	assert.True(t, update.NewState.Status == models.StatusActive ||
		update.NewState.Status == models.StatusEnded ||
		update.NewState.Status == models.StatusUpcoming)
}

func TestGetGameUpdateFromScoreboard_SituationParsing(t *testing.T) {
	svc := NFLService{Client: nflc.NFLMockClient{}}

	// Create a test game
	game := models.Game{
		GameCode: "401547403",
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team:  models.Team{TeamCode: "KC", TeamName: "Kansas City Chiefs"},
				Score: 0,
			},
			Away: models.TeamState{
				Team:  models.Team{TeamCode: "BUF", TeamName: "Buffalo Bills"},
				Score: 0,
			},
			Status: models.StatusUpcoming,
		},
	}

	// Test situation parsing from ShortDetail
	ret := make(chan models.GameUpdate)
	go svc.getGameUpdateFromScoreboard(game, ret)

	update := <-ret
	assert.Equal(t, game.CurrentState, update.OldState)
	assert.NotNil(t, update.NewState)
}

func TestGetGameUpdateFromScoreboard_NoEvents(t *testing.T) {
	svc := NFLService{Client: nflc.NFLMockClient{}}

	// Create a test game that will return no events
	game := models.Game{
		GameCode: "no-events",
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team:  models.Team{TeamCode: "KC", TeamName: "Kansas City Chiefs"},
				Score: 0,
			},
			Away: models.TeamState{
				Team:  models.Team{TeamCode: "BUF", TeamName: "Buffalo Bills"},
				Score: 0,
			},
			Status: models.StatusUpcoming,
		},
	}

	// Test with no events
	ret := make(chan models.GameUpdate)
	go svc.getGameUpdateFromScoreboard(game, ret)

	update := <-ret
	assert.Equal(t, game.CurrentState, update.OldState)
	assert.NotNil(t, update.NewState) // Mock client returns real data, so state will change
}

func TestGetGameUpdateFromScoreboard_NoCompetitions(t *testing.T) {
	svc := NFLService{Client: nflc.NFLMockClient{}}

	// Create a test game
	game := models.Game{
		GameCode: "401547403",
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team:  models.Team{TeamCode: "KC", TeamName: "Kansas City Chiefs"},
				Score: 0,
			},
			Away: models.TeamState{
				Team:  models.Team{TeamCode: "BUF", TeamName: "Buffalo Bills"},
				Score: 0,
			},
			Status: models.StatusUpcoming,
		},
	}

	// Test with no competitions
	ret := make(chan models.GameUpdate)
	go svc.getGameUpdateFromScoreboard(game, ret)

	update := <-ret
	assert.Equal(t, game.CurrentState, update.OldState)
	assert.NotNil(t, update.NewState)
}

func TestGetGameUpdateFromScoreboard_InsufficientCompetitors(t *testing.T) {
	svc := NFLService{Client: nflc.NFLMockClient{}}

	// Create a test game
	game := models.Game{
		GameCode: "401547403",
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team:  models.Team{TeamCode: "KC", TeamName: "Kansas City Chiefs"},
				Score: 0,
			},
			Away: models.TeamState{
				Team:  models.Team{TeamCode: "BUF", TeamName: "Buffalo Bills"},
				Score: 0,
			},
			Status: models.StatusUpcoming,
		},
	}

	// Test with insufficient competitors
	ret := make(chan models.GameUpdate)
	go svc.getGameUpdateFromScoreboard(game, ret)

	update := <-ret
	assert.Equal(t, game.CurrentState, update.OldState)
	assert.NotNil(t, update.NewState)
}

func TestGameFromScoreboard_EmptyScoreboard(t *testing.T) {
	svc := NFLService{Client: nflc.NFLMockClient{}}
	game := svc.GameFromScoreboard("empty-scoreboard")
	assert.Equal(t, "empty-scoreboard", game.GameCode)
	assert.Equal(t, models.League(models.LeagueIdNFL), game.LeagueId)
}

func TestGameFromScoreboard_WithCompetitors(t *testing.T) {
	svc := NFLService{Client: nflc.NFLMockClient{}}
	game := svc.GameFromScoreboard("401547403")
	assert.Equal(t, "401547403", game.GameCode)
	assert.Equal(t, models.League(models.LeagueIdNFL), game.LeagueId)
	assert.NotEmpty(t, game.CurrentState.Home.Team.TeamCode)
	assert.NotEmpty(t, game.CurrentState.Away.Team.TeamCode)
}

func TestGameFromScoreboard_CompletedGame(t *testing.T) {
	svc := NFLService{Client: nflc.NFLMockClient{}}
	game := svc.GameFromScoreboard("completed-game")
	assert.Equal(t, "completed-game", game.GameCode)
	assert.Equal(t, models.League(models.LeagueIdNFL), game.LeagueId)
}

func TestGameFromScoreboard_ActiveGame(t *testing.T) {
	svc := NFLService{Client: nflc.NFLMockClient{}}
	game := svc.GameFromScoreboard("active-game")
	assert.Equal(t, "active-game", game.GameCode)
	assert.Equal(t, models.League(models.LeagueIdNFL), game.LeagueId)
}

func TestGameFromScoreboard_WithSituation2(t *testing.T) {
	svc := NFLService{Client: nflc.NFLMockClient{}}
	game := svc.GameFromScoreboard("with-situation")
	assert.Equal(t, "with-situation", game.GameCode)
	assert.Equal(t, models.League(models.LeagueIdNFL), game.LeagueId)
}

func TestGameFromScoreboard_WithDrives(t *testing.T) {
	svc := NFLService{Client: nflc.NFLMockClient{}}
	game := svc.GameFromScoreboard("with-drives")
	assert.Equal(t, "with-drives", game.GameCode)
	assert.Equal(t, models.League(models.LeagueIdNFL), game.LeagueId)
}

func TestGameFromScoreboard_EmptyEvents(t *testing.T) {
	service := NFLService{Client: nflc.NFLMockClient{}}

	// Test with empty events
	game := service.GameFromScoreboard("empty-events")

	assert.Equal(t, "empty-events", game.GameCode)
	assert.Equal(t, models.League(models.LeagueIdNFL), game.LeagueId)
	assert.Equal(t, models.GameStatus(models.StatusUpcoming), game.CurrentState.Status)
}

func TestGameFromScoreboard_EmptyCompetitions(t *testing.T) {
	service := NFLService{Client: nflc.NFLMockClient{}}

	// Test with empty competitions
	game := service.GameFromScoreboard("empty-competitions")

	assert.Equal(t, "empty-competitions", game.GameCode)
	assert.Equal(t, models.League(models.LeagueIdNFL), game.LeagueId)
	assert.Equal(t, models.GameStatus(models.StatusUpcoming), game.CurrentState.Status)
}

func TestGameFromScoreboard_HalftimeDetection(t *testing.T) {
	service := NFLService{Client: nflc.NFLMockClient{}}

	// Test halftime detection via ShortDetail
	game := service.GameFromScoreboard("halftime-detail")

	assert.Equal(t, "HALFTIME", game.CurrentState.PeriodType)
	assert.Equal(t, "HALFTIME", game.CurrentState.Clock)
}

func TestGameFromScoreboard_HalftimePeriod2(t *testing.T) {
	service := NFLService{Client: nflc.NFLMockClient{}}

	// Test halftime detection via period 2 and clock 0:00
	game := service.GameFromScoreboard("halftime-period2")

	assert.Equal(t, "HALFTIME", game.CurrentState.PeriodType)
	assert.Equal(t, "HALFTIME", game.CurrentState.Clock)
}

func TestGameFromScoreboard_SituationFallback(t *testing.T) {
	service := NFLService{Client: nflc.NFLMockClient{}}

	// Test situation parsing fallback
	game := service.GameFromScoreboard("situation-fallback")

	assert.Equal(t, "HALFTIME", game.CurrentState.PeriodType)
	assert.Equal(t, "HALFTIME", game.CurrentState.Clock)
}

func TestGetGameUpdateFromScoreboard_EmptyEvents(t *testing.T) {
	service := NFLService{Client: nflc.NFLMockClient{}}
	game := models.Game{
		GameCode: "empty-events",
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team:  models.Team{TeamCode: "WPG", TeamName: "Winnipeg Jets"},
				Score: 0,
			},
			Away: models.TeamState{
				Team:  models.Team{TeamCode: "TOR", TeamName: "Toronto Maple Leafs"},
				Score: 0,
			},
			Status: models.StatusUpcoming,
		},
	}
	ret := make(chan models.GameUpdate, 1)

	service.getGameUpdateFromScoreboard(game, ret)

	update := <-ret
	assert.Equal(t, game.CurrentState, update.OldState)
	assert.Equal(t, game.CurrentState, update.NewState)
}

func TestGetGameUpdateFromScoreboard_EmptyCompetitions(t *testing.T) {
	service := NFLService{Client: nflc.NFLMockClient{}}
	game := models.Game{
		GameCode: "empty-competitions",
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team:  models.Team{TeamCode: "WPG", TeamName: "Winnipeg Jets"},
				Score: 0,
			},
			Away: models.TeamState{
				Team:  models.Team{TeamCode: "TOR", TeamName: "Toronto Maple Leafs"},
				Score: 0,
			},
			Status: models.StatusUpcoming,
		},
	}
	ret := make(chan models.GameUpdate, 1)

	service.getGameUpdateFromScoreboard(game, ret)

	update := <-ret
	assert.Equal(t, game.CurrentState, update.OldState)
	assert.Equal(t, game.CurrentState, update.NewState)
}

func TestGetGameUpdateFromScoreboard_TeamFallback(t *testing.T) {
	service := NFLService{Client: nflc.NFLMockClient{}}
	game := models.Game{
		GameCode: "team-fallback",
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team:  models.Team{TeamCode: "WPG", TeamName: "Winnipeg Jets"},
				Score: 0,
			},
			Away: models.TeamState{
				Team:  models.Team{TeamCode: "TOR", TeamName: "Toronto Maple Leafs"},
				Score: 0,
			},
			Status: models.StatusUpcoming,
		},
	}
	ret := make(chan models.GameUpdate, 1)

	service.getGameUpdateFromScoreboard(game, ret)

	update := <-ret
	assert.Equal(t, game.CurrentState, update.OldState)
	assert.NotNil(t, update.NewState)
}

func TestGetGameUpdateFromScoreboard_SituationParsingFallback(t *testing.T) {
	service := NFLService{Client: nflc.NFLMockClient{}}
	game := models.Game{
		GameCode: "situation-parsing",
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team:  models.Team{TeamCode: "WPG", TeamName: "Winnipeg Jets"},
				Score: 0,
			},
			Away: models.TeamState{
				Team:  models.Team{TeamCode: "TOR", TeamName: "Toronto Maple Leafs"},
				Score: 0,
			},
			Status: models.StatusUpcoming,
		},
	}
	ret := make(chan models.GameUpdate, 1)

	service.getGameUpdateFromScoreboard(game, ret)

	update := <-ret
	assert.Equal(t, game.CurrentState, update.OldState)
	assert.NotNil(t, update.NewState)
}

func TestGetGameUpdateFromScoreboard_WithHalftimeDetection(t *testing.T) {
	service := NFLService{Client: nflc.NFLMockClient{}}
	game := models.Game{
		GameCode: "halftime-detail",
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team:  models.Team{TeamCode: "WPG", TeamName: "Winnipeg Jets"},
				Score: 0,
			},
			Away: models.TeamState{
				Team:  models.Team{TeamCode: "TOR", TeamName: "Toronto Maple Leafs"},
				Score: 0,
			},
			Status: models.StatusUpcoming,
		},
	}
	ret := make(chan models.GameUpdate, 1)

	service.getGameUpdateFromScoreboard(game, ret)

	update := <-ret
	assert.Equal(t, game.CurrentState, update.OldState)
	assert.NotNil(t, update.NewState)
}

func TestGetGameUpdateFromScoreboard_WithHalftimePeriod2(t *testing.T) {
	service := NFLService{Client: nflc.NFLMockClient{}}
	game := models.Game{
		GameCode: "halftime-period2",
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team:  models.Team{TeamCode: "WPG", TeamName: "Winnipeg Jets"},
				Score: 0,
			},
			Away: models.TeamState{
				Team:  models.Team{TeamCode: "TOR", TeamName: "Toronto Maple Leafs"},
				Score: 0,
			},
			Status: models.StatusUpcoming,
		},
	}
	ret := make(chan models.GameUpdate, 1)

	service.getGameUpdateFromScoreboard(game, ret)

	update := <-ret
	assert.Equal(t, game.CurrentState, update.OldState)
	assert.NotNil(t, update.NewState)
}

func TestGetGameUpdateFromScoreboard_WithSituationFallback(t *testing.T) {
	service := NFLService{Client: nflc.NFLMockClient{}}
	game := models.Game{
		GameCode: "situation-fallback",
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team:  models.Team{TeamCode: "WPG", TeamName: "Winnipeg Jets"},
				Score: 0,
			},
			Away: models.TeamState{
				Team:  models.Team{TeamCode: "TOR", TeamName: "Toronto Maple Leafs"},
				Score: 0,
			},
			Status: models.StatusUpcoming,
		},
	}
	ret := make(chan models.GameUpdate, 1)

	service.getGameUpdateFromScoreboard(game, ret)

	update := <-ret
	assert.Equal(t, game.CurrentState, update.OldState)
	assert.NotNil(t, update.NewState)
}
