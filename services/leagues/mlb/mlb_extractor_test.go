package mlb

import (
	mlb "goalfeed/clients/leagues/mlb"
	"goalfeed/models"
	"testing"
)

func TestExtractCurrentPitcherAndBatter(t *testing.T) {
	svc := MLBService{}
	teams := mlb.BoxscoreTeams{
		Home: mlb.TeamBoxscore{Players: map[string]mlb.Player{
			"ID1": {Person: mlb.Person{ID: 1, FullName: "Pitcher One"}, JerseyNumber: "45", Position: mlb.Position{Name: "P"}, GameStatus: mlb.GameStatus{IsCurrentPitcher: true}},
			"ID2": {Person: mlb.Person{ID: 2, FullName: "Batter One"}, JerseyNumber: "12", Position: mlb.Position{Name: "1B"}, GameStatus: mlb.GameStatus{IsCurrentBatter: true}},
		}},
		Away: mlb.TeamBoxscore{Players: map[string]mlb.Player{}},
	}
	p := svc.extractCurrentPitcher(teams)
	if p.Name != "Pitcher One" || p.Number != 45 || p.Position != "P" {
		t.Fatalf("unexpected pitcher: %+v", p)
	}
	b := svc.extractCurrentBatter(teams)
	if b.Name != "Batter One" || b.Number != 12 || b.Position != "1B" {
		t.Fatalf("unexpected batter: %+v", b)
	}
}

func TestExtractorsWhenNoneFound(t *testing.T) {
	svc := MLBService{}
	empty := mlb.BoxscoreTeams{Home: mlb.TeamBoxscore{Players: map[string]mlb.Player{}}, Away: mlb.TeamBoxscore{Players: map[string]mlb.Player{}}}
	p := svc.extractCurrentPitcher(empty)
	if p != (models.Player{}) {
		t.Fatalf("expected zero pitcher")
	}
	b := svc.extractCurrentBatter(empty)
	if b != (models.Player{}) {
		t.Fatalf("expected zero batter")
	}
}
