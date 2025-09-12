package homeassistant

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"goalfeed/models"
)

func withHAServer(t *testing.T) *httptest.Server {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	os.Setenv("SUPERVISOR_API", srv.URL)
	os.Setenv("SUPERVISOR_TOKEN", "t")
	t.Cleanup(func() {
		os.Unsetenv("SUPERVISOR_API")
		os.Unsetenv("SUPERVISOR_TOKEN")
	})
	return srv
}

func TestToStateString(t *testing.T) {
	if toStateString(true) != "on" || toStateString(false) != "off" {
		t.Fatal("bool mapping failed")
	}
	if toStateString(3) != "3" {
		t.Fatal("int mapping failed")
	}
	if toStateString(3.5) != "3.5" {
		t.Fatal("float mapping failed")
	}
	if toStateString("") != "" {
		t.Fatal("string passthrough failed")
	}
	if toStateString(time.Time{}) != "unknown" {
		t.Fatal("zero time should be unknown")
	}
}

func TestPublishEntityCachesAndSends(t *testing.T) {
	withHAServer(t)
	// Avoid debounce interfering with test
	debounceAfter = 0
	ok, prev := publishEntity("sensor", "goalfeed_nhl_wpg_team_status", "WPG team status", "active", nil)
	if !ok || prev != "" {
		t.Fatalf("first publish should send, prev empty; got ok=%v prev=%q", ok, prev)
	}
	// duplicate should be deduped
	ok2, _ := publishEntity("sensor", "goalfeed_nhl_wpg_team_status", "WPG team status", "active", nil)
	if ok2 {
		t.Fatalf("duplicate publish should be deduped")
	}
}

func TestBuildEntityNameAndSlug(t *testing.T) {
	if leagueSlug(models.LeagueIdNFL) != "nfl" {
		t.Fatal("slug nfl failed")
	}
	name := buildEntityName(models.LeagueIdNHL, "WPG", "team.status")
	if name != "goalfeed_nhl_wpg_team_status" {
		t.Fatalf("unexpected entity name: %s", name)
	}
	if sanitizeId("Goalfeed NHL WPG-Team#1") != "goalfeed_nhl_wpg_team_1" {
		t.Fatal("sanitize failed")
	}
}

func TestPublishScheduleAndEndOfGame(t *testing.T) {
	withHAServer(t)
	game := models.Game{
		LeagueId: models.LeagueIdNHL,
		GameDetails: models.GameDetails{
			GameDate: time.Now(),
		},
		CurrentState: models.GameState{
			Status: models.StatusUpcoming,
			Home:   models.TeamState{Team: models.Team{TeamCode: "WPG"}},
			Away:   models.TeamState{Team: models.Team{TeamCode: "EDM"}},
		},
	}
	PublishScheduleSensorsForGame(game)
	// End of game resets
	game.CurrentState.Status = models.StatusEnded
	PublishEndOfGameReset(game)
}

func TestPublishTeamSensorsNHL(t *testing.T) {
	withHAServer(t)
	debounceAfter = 0
	game := models.Game{
		LeagueId: models.LeagueIdNHL,
		CurrentState: models.GameState{
			Status: models.StatusActive,
			Period: 2,
			Clock:  "05:00",
			Home:   models.TeamState{Team: models.Team{TeamCode: "WPG"}, Score: 2, Statistics: models.TeamStats{Shots: 20, Penalties: 2}},
			Away:   models.TeamState{Team: models.Team{TeamCode: "EDM"}, Score: 1, Statistics: models.TeamStats{Shots: 15, Penalties: 1}},
		},
	}
	PublishTeamSensors(game)
}

func TestPublishTeamSensorsNFL(t *testing.T) {
	withHAServer(t)
	debounceAfter = 0
	game := models.Game{
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Status:  models.StatusActive,
			Period:  3,
			Clock:   "07:12",
			Details: models.EventDetails{Possession: "BUF", Down: 2, Distance: 6, YardLine: 85},
			Home:    models.TeamState{Team: models.Team{TeamCode: "MIA"}, Score: 10},
			Away:    models.TeamState{Team: models.Team{TeamCode: "BUF"}, Score: 17},
		},
	}
	PublishTeamSensors(game)
}

func TestPublishTeamSensorsMLB(t *testing.T) {
	withHAServer(t)
	debounceAfter = 0
	game := models.Game{
		LeagueId: models.LeagueIdMLB,
		CurrentState: models.GameState{
			Status:  models.StatusActive,
			Details: models.EventDetails{Possession: "TOR", BallCount: 2, StrikeCount: 1, Outs: 1, Bases: "1st", Pitcher: models.Player{Name: "P"}, Batter: models.Player{Name: "B"}},
			Home:    models.TeamState{Team: models.Team{TeamCode: "NYY"}, Score: 3},
			Away:    models.TeamState{Team: models.Team{TeamCode: "TOR"}, Score: 4},
		},
	}
	PublishTeamSensors(game)
}

func TestStatusString(t *testing.T) {
	if statusString(models.StatusUpcoming) != "scheduled" {
		t.Fatal("upcoming")
	}
	if statusString(models.StatusActive) != "active" {
		t.Fatal("active")
	}
	if statusString(models.StatusDelayed) != "delayed" {
		t.Fatal("delayed")
	}
	if statusString(models.StatusEnded) != "final" {
		t.Fatal("ended")
	}
	if statusString(999) != "unknown" {
		t.Fatal("unknown")
	}
}

func TestPublishEntityNoHAConfigured(t *testing.T) {
	// Ensure no env set
	os.Unsetenv("SUPERVISOR_API")
	os.Unsetenv("SUPERVISOR_TOKEN")
	debounceAfter = 0
	ok, _ := publishEntity("sensor", "goalfeed_nhl_wpg_test", "", "x", nil)
	if ok {
		t.Fatalf("expected no send when HA not configured")
	}
}
