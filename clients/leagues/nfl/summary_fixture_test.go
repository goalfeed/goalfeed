package nfl

import (
	"encoding/json"
	"os"
	"testing"
)

// testdata/nfl_summary_completed_401873286.json is a REAL capture of
// ESPN's "/apis/site/v2/sports/football/nfl/summary?event=401873286"
// endpoint (LV @ HOU), fetched post-game on 2026-08-21, trimmed to the
// header.competitions subtree plus gameInfo.venue -- i.e. exactly what
// GetNFLScoreBoard's callers read (see extractSummarySnapshot in
// services/leagues/nfl). The trim drops team logos/links, the notes/
// broadcasts noise, and the (large, unrelated) boxscore/leaders/injuries/
// odds/videos/standings sections; every field this package's parsing
// reads is byte-for-byte as captured. Confirms the real payload has NO
// top-level "events" key at all -- the bug this fixture exists to catch.
func TestNFLSummary_ParsesScoreFromHeaderCompetitions(t *testing.T) {
	b, err := os.ReadFile("testdata/nfl_summary_completed_401873286.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var resp NFLScoreboardResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		t.Fatalf("unmarshal real summary fixture: %v", err)
	}

	if len(resp.Events) != 0 {
		t.Fatalf("expected the real summary payload to have NO events array, got %d events", len(resp.Events))
	}
	if len(resp.Header.Competitions) != 1 {
		t.Fatalf("expected 1 header competition, got %d", len(resp.Header.Competitions))
	}

	comp := resp.Header.Competitions[0]
	if len(comp.Competitors) != 2 {
		t.Fatalf("expected 2 competitors, got %d", len(comp.Competitors))
	}

	var home, away NFLScoreboardCompetitor
	for _, c := range comp.Competitors {
		if c.HomeAway == "home" {
			home = c
		} else if c.HomeAway == "away" {
			away = c
		}
	}

	if home.Team.Abbreviation != "HOU" || home.Score != "20" {
		t.Errorf("expected home HOU 20, got %s %s", home.Team.Abbreviation, home.Score)
	}
	if away.Team.Abbreviation != "LV" || away.Score != "22" {
		t.Errorf("expected away LV 22, got %s %s", away.Team.Abbreviation, away.Score)
	}
	if !comp.Status.Type.Completed || comp.Status.Type.State != "post" {
		t.Errorf("expected completed/post status, got completed=%v state=%s", comp.Status.Type.Completed, comp.Status.Type.State)
	}
	if resp.GameInfo.Venue.FullName != "Reliant Stadium" {
		t.Errorf("expected venue Reliant Stadium, got %q", resp.GameInfo.Venue.FullName)
	}
	if resp.Header.Week != 3 {
		t.Errorf("expected week 3, got %d", resp.Header.Week)
	}
}
