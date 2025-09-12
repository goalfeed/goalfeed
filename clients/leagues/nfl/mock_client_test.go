package nfl

import "testing"

func TestNFLMockClient_Getters(t *testing.T) {
	c := NFLMockClient{}

	sched := c.GetNFLSchedule()
	if len(sched.Events) == 0 {
		t.Fatalf("expected mocked schedule events")
	}

	sb := c.GetNFLScoreBoard("401547403")
	if len(sb.Events) == 0 {
		t.Fatalf("expected mocked scoreboard events")
	}

	team := c.GetTeam("BUF")
	if len(team.Teams) == 0 || team.Teams[0].Abbreviation != "BUF" {
		t.Fatalf("expected mocked team BUF, got %+v", team)
	}

	all := c.GetAllTeams()
	if len(all.Teams) == 0 {
		t.Fatalf("expected mocked teams list")
	}
}
