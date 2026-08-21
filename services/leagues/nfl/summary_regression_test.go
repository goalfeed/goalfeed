package nfl

import (
	"encoding/json"
	"os"
	"testing"

	nflc "goalfeed/clients/leagues/nfl"
	"goalfeed/models"
)

// Fixture provenance (duplicated from clients/leagues/nfl/testdata, which
// carries the full provenance comment):
//
//   - testdata/nfl_summary_completed_401873286.json is a REAL capture of
//     ESPN's "/summary?event=401873286" endpoint (LV @ HOU), fetched
//     post-game on 2026-08-21, trimmed to header.competitions +
//     gameInfo.venue.
//   - testdata/nfl_summary_before_score_change.json is HAND-DERIVED from
//     that real fixture: the away (LV) score was edited down from "22" to
//     "15" (and the corresponding 4th-quarter linescore from 12 to 5), and
//     the status block was replaced with an in-progress snapshot (period
//     4, displayClock "9:12", state "in", completed false) to simulate the
//     moment before LV's go-ahead score. No other field was changed. This
//     is the only fixture in this pair that isn't a raw capture -- it
//     exists purely to give getGameUpdateFromScoreboard a "before" state
//     to diff against the real "after" fixture.
func loadNFLFixture(t *testing.T, name string) nflc.NFLScoreboardResponse {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", name, err)
	}
	var resp nflc.NFLScoreboardResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		t.Fatalf("unmarshal fixture %s: %v", name, err)
	}
	return resp
}

// fixtureNFLClient serves a fixed GetNFLScoreBoard response, embedding
// NFLMockClient so the other INFLAPIClient methods (unused by this test)
// still have sane implementations.
type fixtureNFLClient struct {
	nflc.NFLMockClient
	resp nflc.NFLScoreboardResponse
}

func (f fixtureNFLClient) GetNFLScoreBoard(gameId string) nflc.NFLScoreboardResponse {
	return f.resp
}

// TestNFLGetGameUpdate_DetectsScoreChange is the regression test that pins
// down the actual production bug: GetNFLScoreBoard's real HTTP response
// (the "/summary?event=" shape) has no top-level "events" array, so
// getGameUpdateFromScoreboard used to find len(scoreboard.Events) == 0
// every time, never enter its update block, and return NewState identical
// to OldState -- meaning GetEvents' `diff := newState.Score -
// oldState.Score` was always 0 and no goal/score event could ever fire.
//
// This test builds the "old" game state from a hand-derived pre-score
// fixture (LV 15) and asks for an update using the real captured "after"
// fixture (LV 22), then asserts the new state actually reflects the higher
// score. Against the pre-fix code (matching on scoreboard.Events instead of
// scoreboard.Header), this fails because NewState stays at the old score.
func TestNFLGetGameUpdate_DetectsScoreChange(t *testing.T) {
	before := loadNFLFixture(t, "nfl_summary_before_score_change.json")
	after := loadNFLFixture(t, "nfl_summary_completed_401873286.json")

	// Build the starting game state from the "before" fixture.
	svcBefore := NFLService{Client: fixtureNFLClient{resp: before}}
	startGame := svcBefore.GameFromScoreboard("401873286")
	if startGame.CurrentState.Away.Score != 15 {
		t.Fatalf("sanity check failed: expected starting away score 15, got %d", startGame.CurrentState.Away.Score)
	}
	if startGame.CurrentState.Home.Score != 20 {
		t.Fatalf("sanity check failed: expected starting home score 20, got %d", startGame.CurrentState.Home.Score)
	}

	// Now ask for an update using the real "after" fixture as the live
	// response.
	svcAfter := NFLService{Client: fixtureNFLClient{resp: after}}
	ch := make(chan models.GameUpdate)
	go svcAfter.getGameUpdateFromScoreboard(startGame, ch)
	update := <-ch

	if update.OldState.Away.Score != 15 {
		t.Fatalf("expected OldState away score to be preserved as 15, got %d", update.OldState.Away.Score)
	}
	if update.NewState.Away.Score != 22 {
		t.Fatalf("expected NewState away score 22 (the real captured score), got %d -- score change was not detected", update.NewState.Away.Score)
	}
	if update.NewState.Home.Score != 20 {
		t.Fatalf("expected NewState home score 20, got %d", update.NewState.Home.Score)
	}
	if update.NewState.Status != models.StatusEnded {
		t.Fatalf("expected NewState status Ended (real fixture is Final), got %v", update.NewState.Status)
	}

	// The consequence check called out in the bug report: GetEvents must
	// actually fire for this real score change.
	svc := NFLService{}
	evCh := make(chan []models.Event)
	go svc.GetEvents(update, evCh)
	events := <-evCh
	if len(events) == 0 {
		t.Fatalf("expected at least one event for a 7-point away score increase, got none")
	}
	for _, ev := range events {
		if ev.TeamCode != "LV" {
			t.Fatalf("expected scoring events for LV, got %s", ev.TeamCode)
		}
	}
}
