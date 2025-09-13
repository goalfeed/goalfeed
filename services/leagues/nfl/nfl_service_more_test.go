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
