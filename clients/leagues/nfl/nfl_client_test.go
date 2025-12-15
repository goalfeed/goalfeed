package nfl

import (
	"encoding/json"
	"testing"
)

func withStubFetch(t *testing.T, payload interface{}) func() {
	old := fetchByte
	b, _ := json.Marshal(payload)
	fetchByte = func(url string, ret chan []byte) {
		ret <- b
	}
	return func() { fetchByte = old }
}

func TestNFLClient_Schedule(t *testing.T) {
	restore := withStubFetch(t, NFLScheduleResponse{Events: []NFLScheduleEvent{{ID: "1"}}})
	defer restore()
	c := NFLAPIClient{}
	resp := c.GetNFLSchedule()
	if len(resp.Events) != 1 || resp.Events[0].ID != "1" {
		t.Fatalf("unexpected schedule resp: %+v", resp)
	}
}

func TestNFLClient_Scoreboard(t *testing.T) {
	restore := withStubFetch(t, NFLScoreboardResponse{Events: []NFLScoreboardEvent{{ID: "401"}}})
	defer restore()
	c := NFLAPIClient{}
	resp := c.GetNFLScoreBoard("401")
	if len(resp.Events) != 1 || resp.Events[0].ID != "401" {
		t.Fatalf("unexpected scoreboard resp: %+v", resp)
	}
}

func TestNFLClient_Team(t *testing.T) {
	restore := withStubFetch(t, NFLTeamResponse{Teams: []NFLTeam{{ID: "4", DisplayName: "Buffalo Bills"}}})
	defer restore()
	c := NFLAPIClient{}
	resp := c.GetTeam("BUF")
	if len(resp.Teams) != 1 || resp.Teams[0].ID != "4" {
		t.Fatalf("unexpected team resp: %+v", resp)
	}
}

func TestNFLClient_AllTeams(t *testing.T) {
	restore := withStubFetch(t, NFLTeamResponse{Teams: []NFLTeam{{ID: "4"}, {ID: "15"}}})
	defer restore()
	c := NFLAPIClient{}
	resp := c.GetAllTeams()
	if len(resp.Teams) != 2 {
		t.Fatalf("unexpected all teams resp: %+v", resp)
	}
}
