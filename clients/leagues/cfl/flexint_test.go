package cfl

import (
	"encoding/json"
	"os"
	"testing"
)

// Fixture provenance:
//
//   - testdata/cfl_live_game_stringid.json is a REAL capture of the
//     BetGenius live-game endpoint (multisportgametracker) for CFL fixture
//     13419716 (MTL @ OTT), fetched post-game on 2026-08-21. It is trimmed
//     to keep only the first play-by-play "matchActions" entry (the parser
//     doesn't read that array at all); every field the parser DOES read
//     (sportId/competitionId as strings, scoreboardInfo, matchInfo,
//     availableTabs) is byte-for-byte as captured. This is the exact shape
//     that used to make GetCFLLiveGame discard the whole response: sportId
//     and competitionId arrive as JSON strings ("17"/"1035"), not numbers.
//   - testdata/cfl_live_game_numericid.json is that same real capture with
//     ONLY sportId/competitionId hand-edited from strings to JSON numbers
//     (17/1035), to prove FlexInt accepts both encodings.
func loadCFLFixture(t *testing.T, name string) CFLLiveGameResponse {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", name, err)
	}
	var resp CFLLiveGameResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		t.Fatalf("unmarshal fixture %s: %v", name, err)
	}
	return resp
}

// TestCFLLiveGameResponse_ParsesStringTypedIDs pins down the actual
// regression: a live BetGenius payload sends sportId/competitionId as JSON
// strings. Before FlexInt, unmarshaling this fixture into
// CFLLiveGameResponse failed with "cannot unmarshal string into Go struct
// field CFLLiveGameResponse.sportId of type int", and the caller
// (GetCFLLiveGame) treated that error as fatal and returned a zero struct
// -- discarding the correctly-typed score fields right along with it.
func TestCFLLiveGameResponse_ParsesStringTypedIDs(t *testing.T) {
	resp := loadCFLFixture(t, "cfl_live_game_stringid.json")

	if resp.SportID != 17 {
		t.Errorf("expected SportID 17, got %d", resp.SportID)
	}
	if resp.CompetitionID != 1035 {
		t.Errorf("expected CompetitionID 1035, got %d", resp.CompetitionID)
	}
	if resp.Data.ScoreboardInfo.HomeScore != 46 {
		t.Errorf("expected HomeScore 46, got %d", resp.Data.ScoreboardInfo.HomeScore)
	}
	if resp.Data.ScoreboardInfo.AwayScore != 16 {
		t.Errorf("expected AwayScore 16, got %d", resp.Data.ScoreboardInfo.AwayScore)
	}
	if resp.Data.ScoreboardInfo.MatchStatus != "PostMatch" {
		t.Errorf("expected MatchStatus PostMatch, got %q", resp.Data.ScoreboardInfo.MatchStatus)
	}
}

// TestCFLLiveGameResponse_ParsesNumberTypedIDs proves FlexInt also handles
// a plain numeric encoding, so a future upstream flip back to real JSON
// numbers for sportId/competitionId won't break parsing again.
func TestCFLLiveGameResponse_ParsesNumberTypedIDs(t *testing.T) {
	resp := loadCFLFixture(t, "cfl_live_game_numericid.json")

	if resp.SportID != 17 {
		t.Errorf("expected SportID 17, got %d", resp.SportID)
	}
	if resp.CompetitionID != 1035 {
		t.Errorf("expected CompetitionID 1035, got %d", resp.CompetitionID)
	}
	if resp.Data.ScoreboardInfo.HomeScore != 46 {
		t.Errorf("expected HomeScore 46, got %d", resp.Data.ScoreboardInfo.HomeScore)
	}
	if resp.Data.ScoreboardInfo.AwayScore != 16 {
		t.Errorf("expected AwayScore 16, got %d", resp.Data.ScoreboardInfo.AwayScore)
	}
}

// TestGetCFLLiveGame_RealPayloadNotDiscarded exercises the actual client
// method (not just json.Unmarshal directly) against the real string-typed
// fixture, proving the full GetCFLLiveGame path -- including the
// log-and-keep-partial-data behavior on any future unmarshal error --
// returns real score data rather than a zero struct.
func TestGetCFLLiveGame_RealPayloadNotDiscarded(t *testing.T) {
	b, err := os.ReadFile("testdata/cfl_live_game_stringid.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}
	oldB := fetchByte
	defer func() { fetchByte = oldB }()
	fetchByte = func(url string, ret chan []byte) { ret <- b }

	c := CFLApiClient{}
	resp := c.GetCFLLiveGame("13419716")

	if resp.SportID != 17 || resp.CompetitionID != 1035 {
		t.Fatalf("expected SportID/CompetitionID to parse, got %d/%d", resp.SportID, resp.CompetitionID)
	}
	if resp.Data.ScoreboardInfo.HomeScore != 46 || resp.Data.ScoreboardInfo.AwayScore != 16 {
		t.Fatalf("expected real score 46-16, got %d-%d", resp.Data.ScoreboardInfo.HomeScore, resp.Data.ScoreboardInfo.AwayScore)
	}
}
