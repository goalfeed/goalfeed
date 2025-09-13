package nfl

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"goalfeed/models"
	"goalfeed/targets/memoryStore"

	"github.com/spf13/viper"
)

func seedNFLGame(eventID string) models.Game {
	g := models.Game{
		GameCode: eventID,
		LeagueId: models.LeagueIdNFL,
		CurrentState: models.GameState{
			Home:   models.TeamState{Team: models.Team{TeamCode: "MIA", TeamName: "Miami", ExtID: "15", LeagueID: models.LeagueIdNFL}},
			Away:   models.TeamState{Team: models.Team{TeamCode: "BUF", TeamName: "Buffalo", ExtID: "4", LeagueID: models.LeagueIdNFL}},
			Status: models.StatusUpcoming,
			Clock:  "",
			Period: 0,
		},
	}
	memoryStore.SetGame(g)
	memoryStore.AppendActiveGame(g)
	return g
}

func TestApplyNFLPatches_DirectArray(t *testing.T) {
	memoryStore.ClearAllGames()
	eventID := "401547403"
	seedNFLGame(eventID)
	ops := []patchOp{
		{Op: "replace", Path: "/fullStatus/displayClock", Value: "6:14"},
		{Op: "replace", Path: "/fullStatus/type/shortDetail", Value: "6:14 - 3rd"},
		{Op: "replace", Path: "/situation/down", Value: float64(2)},
		{Op: "replace", Path: "/situation/yardLine", Value: float64(35)},
		{Op: "replace", Path: "/possessionText", Value: "BUF 35"},
		{Op: "replace", Path: "/situation/shortDownDistanceText", Value: "2nd & 8 at BUF 35"},
		{Op: "replace", Path: "/competitors/0/homeAway", Value: "away"},
		{Op: "replace", Path: "/competitors/0/team/id", Value: "4"},
		{Op: "replace", Path: "/competitors/0/score", Value: "21"},
		{Op: "replace", Path: "/competitors/1/homeAway", Value: "home"},
		{Op: "replace", Path: "/competitors/1/team/id", Value: "15"},
		{Op: "replace", Path: "/competitors/1/score", Value: "17"},
	}
	b, _ := json.Marshal(ops)
	applyNFLPatches(json.RawMessage(b), "gp-football-nfl-"+eventID)
	g, err := memoryStore.GetGameByGameKey(models.Game{GameCode: eventID, LeagueId: models.LeagueIdNFL}.GetGameKey())
	if err != nil {
		t.Fatalf("game not found: %v", err)
	}
	if g.CurrentState.Clock != "6:14" || g.CurrentState.Period != 3 || g.CurrentState.PeriodType != "QUARTER" {
		t.Fatalf("expected period/clock set, got period=%d type=%s clock=%s", g.CurrentState.Period, g.CurrentState.PeriodType, g.CurrentState.Clock)
	}
	if g.CurrentState.Details.Down != 2 || g.CurrentState.Details.YardLine != 35 || g.CurrentState.Details.Possession != "BUF" {
		t.Fatalf("expected situation set, got down=%d yl=%d poss=%s", g.CurrentState.Details.Down, g.CurrentState.Details.YardLine, g.CurrentState.Details.Possession)
	}
	if g.CurrentState.Home.Score != 17 || g.CurrentState.Away.Score != 21 {
		t.Fatalf("expected scores 21-17 (away-home), got %d-%d", g.CurrentState.Away.Score, g.CurrentState.Home.Score)
	}
}

func TestApplyNFLPatches_WrapperBase64Zlib(t *testing.T) {
	memoryStore.ClearAllGames()
	eventID := "401547404"
	seedNFLGame(eventID)
	ops := []patchOp{{Op: "replace", Path: "/fullStatus/type/detail", Value: "3:28 - 3rd Quarter"}}
	plain, _ := json.Marshal(ops)
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	_, _ = zw.Write(plain)
	_ = zw.Close()
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	wrapper := struct {
		Ts int64           `json:"ts"`
		C  int64           `json:"~c"`
		Pl json.RawMessage `json:"pl"`
	}{Ts: time.Now().Unix(), C: 0, Pl: json.RawMessage([]byte("\"" + encoded + "\""))}
	b, _ := json.Marshal(wrapper)
	applyNFLPatches(json.RawMessage(b), "gp-football-nfl-"+eventID)
	g, err := memoryStore.GetGameByGameKey(models.Game{GameCode: eventID, LeagueId: models.LeagueIdNFL}.GetGameKey())
	if err != nil {
		t.Fatalf("game not found: %v", err)
	}
	if g.CurrentState.Period != 3 || g.CurrentState.PeriodType != "QUARTER" || g.CurrentState.Clock != "3:28" {
		t.Fatalf("expected halftime/period mapping from detail, got p=%d type=%s clock=%s", g.CurrentState.Period, g.CurrentState.PeriodType, g.CurrentState.Clock)
	}
}

func TestApplyNFLPatches_FloatClockAndSideScores(t *testing.T) {
	memoryStore.ClearAllGames()
	eventID := "401547405"
	seedNFLGame(eventID)
	ops := []patchOp{
		{Op: "replace", Path: "/fullStatus/clock", Value: float64(360)}, // 6:00
		{Op: "replace", Path: "/competitors/0/homeAway", Value: "home"},
		{Op: "replace", Path: "/competitors/1/homeAway", Value: "away"},
		{Op: "replace", Path: "/competitors/0/score", Value: "10"},
		{Op: "replace", Path: "/competitors/1/score", Value: "7"},
	}
	b, _ := json.Marshal(ops)
	applyNFLPatches(json.RawMessage(b), "gp-football-nfl-"+eventID)
	g, err := memoryStore.GetGameByGameKey(models.Game{GameCode: eventID, LeagueId: models.LeagueIdNFL}.GetGameKey())
	if err != nil {
		t.Fatalf("game not found: %v", err)
	}
	if g.CurrentState.Clock != "6:00" {
		t.Fatalf("expected 6:00 clock from float seconds, got %s", g.CurrentState.Clock)
	}
	if g.CurrentState.Home.Score != 10 || g.CurrentState.Away.Score != 7 {
		t.Fatalf("expected home=10 away=7 via side mapping, got %d-%d", g.CurrentState.Home.Score, g.CurrentState.Away.Score)
	}
}

func TestStartNFLFastcast_Disabled(t *testing.T) {
	viper.Set("nfl.fastcast.enabled", false)
	StartNFLFastcast()
	// Nothing to assert; just cover disabled branch without panic
}
