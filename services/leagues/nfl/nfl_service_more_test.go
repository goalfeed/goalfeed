package nfl

import (
	"testing"

	nflc "goalfeed/clients/leagues/nfl"
	"goalfeed/models"
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
