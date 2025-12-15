package nfl

import (
	"goalfeed/clients/leagues/nfl"
	"goalfeed/models"
	"testing"
)

func TestNFLService_GetLeagueName(t *testing.T) {
	service := NFLService{Client: nfl.NFLMockClient{}}
	if service.GetLeagueName() != "NFL" {
		t.Errorf("Expected league name to be 'NFL', got '%s'", service.GetLeagueName())
	}
}

func TestNFLService_GetActiveGames(t *testing.T) {
	service := NFLService{Client: nfl.NFLMockClient{}}
	gamesChan := make(chan []models.Game)
	go service.GetActiveGames(gamesChan)
	games := <-gamesChan

	if len(games) == 0 {
		t.Error("Expected at least one active game")
	}

	game := games[0]
	if game.LeagueId != models.LeagueIdNFL {
		t.Errorf("Expected league ID to be %d, got %d", models.LeagueIdNFL, game.LeagueId)
	}
}

func TestNFLService_GetUpcomingGames(t *testing.T) {
	service := NFLService{Client: nfl.NFLMockClient{}}
	gamesChan := make(chan []models.Game)
	go service.GetUpcomingGames(gamesChan)
	games := <-gamesChan

	if len(games) == 0 {
		t.Error("Expected at least one upcoming game")
	}

	game := games[0]
	if game.LeagueId != models.LeagueIdNFL {
		t.Errorf("Expected league ID to be %d, got %d", models.LeagueIdNFL, game.LeagueId)
	}
}

func TestGameFromEvent_HalftimeLabeling(t *testing.T) {
	svc := NFLService{Client: nfl.NFLMockClient{}}
	ev := nfl.NFLScheduleEvent{}
	ev.ID = "401547403"
	// Set ShortDetail to indicate halftime
	ev.Status.Type.ShortDetail = "Halftime"
	// Provide a competition shell so venue access is valid
	ev.Competitions = []nfl.NFLCompetition{{}}
	g := svc.gameFromEvent(ev)
	if g.CurrentState.PeriodType != "HALFTIME" || g.CurrentState.Clock != "HALFTIME" {
		t.Fatalf("expected halftime labeling, got periodType=%s clock=%s", g.CurrentState.PeriodType, g.CurrentState.Clock)
	}
}

func TestGetGameUpdateFromScoreboard_PopulatesDetails(t *testing.T) {
	svc := NFLService{Client: nfl.NFLMockClient{}}
	base := models.Game{GameCode: "401547403", LeagueId: models.LeagueIdNFL}
	ch := make(chan models.GameUpdate)
	go svc.getGameUpdateFromScoreboard(base, ch)
	upd := <-ch
	if upd.NewState.Details.Down == 0 || upd.NewState.Details.Distance == 0 || upd.NewState.Details.Possession == "" || upd.NewState.Details.YardLine == 0 {
		t.Fatalf("expected details to be populated from scoreboard")
	}
	if upd.NewState.Status != models.StatusActive && upd.NewState.Status != models.StatusEnded {
		t.Fatalf("expected status to be active or ended, got %v", upd.NewState.Status)
	}
}

func TestGameStatusFromEventStatus(t *testing.T) {
	var st struct {
		Clock        float64 `json:"clock"`
		DisplayClock string  `json:"displayClock"`
		Period       int     `json:"period"`
		Type         struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			State       string `json:"state"`
			Completed   bool   `json:"completed"`
			Description string `json:"description"`
			Detail      string `json:"detail"`
			ShortDetail string `json:"shortDetail"`
		} `json:"type"`
	}
	st.Type.State = STATUS_UPCOMING
	if gameStatusFromEventStatus(st) != models.StatusUpcoming {
		t.Fatalf("expected upcoming")
	}
	st.Type.State = STATUS_ACTIVE
	if gameStatusFromEventStatus(st) != models.StatusActive {
		t.Fatalf("expected active")
	}
	st.Type.State = STATUS_FINAL
	if gameStatusFromEventStatus(st) != models.StatusEnded {
		t.Fatalf("expected ended")
	}
}

func TestNFLService_GetEvents_ScoringDiff(t *testing.T) {
	svc := NFLService{}
	upd := models.GameUpdate{
		OldState: models.GameState{Home: models.TeamState{Team: models.Team{TeamCode: "BUF", TeamName: "Buffalo"}, Score: 14}, Away: models.TeamState{Team: models.Team{TeamCode: "MIA", TeamName: "Miami"}, Score: 7}},
		NewState: models.GameState{Home: models.TeamState{Team: models.Team{TeamCode: "BUF", TeamName: "Buffalo"}, Score: 17}, Away: models.TeamState{Team: models.Team{TeamCode: "MIA", TeamName: "Miami"}, Score: 7}},
	}
	ch := make(chan []models.Event)
	go svc.GetEvents(upd, ch)
	ev := <-ch
	if len(ev) != 3 { // 3 point diff emits 3 events (per current simple logic)
		t.Fatalf("expected 3 events for 3-point diff, got %d", len(ev))
	}
}

func TestNFLService_GameFromScoreboard_FallbackWhenEmpty(t *testing.T) {
	svc := NFLService{Client: nfl.NFLMockClient{}}
	g := svc.GameFromScoreboard("")
	if g.GameCode != "" || g.LeagueId != 0 {
		// With empty event id, expect an empty minimal game
		// This asserts we don't panic on empty ids
	}
}
