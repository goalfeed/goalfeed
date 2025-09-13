package cfl

import (
	"testing"
)

func TestMockCFLApiClient_Getters(t *testing.T) {
	schedule := CFLScheduleResponse{
		{
			ID:     1,
			Status: "scheduled",
			Name:   "Week 1",
			Type:   "REG",
			Number: 1,
			Tournaments: []CFLGame{
				{
					ID:        123,
					Date:      "2025-05-19T20:00:00+00:00",
					HomeSquad: CFLTeam{ID: 100, Name: "Home", ShortName: "HOM"},
					AwaySquad: CFLTeam{ID: 200, Name: "Away", ShortName: "AWY"},
					Status:    "scheduled",
				},
			},
		},
	}

	live := CFLLiveGameResponse{
		Data: CFLLiveGameData{
			BetGeniusFixtureID: "123",
		},
	}

	client := MockCFLApiClient{ScheduleResponse: schedule, LiveGameResponse: live}

	gotSchedule := client.GetCFLSchedule()
	if len(gotSchedule) != 1 || len(gotSchedule[0].Tournaments) != 1 {
		t.Fatalf("expected 1 round with 1 game, got %+v", gotSchedule)
	}
	if gotSchedule[0].Tournaments[0].ID != 123 {
		t.Fatalf("expected game id 123, got %d", gotSchedule[0].Tournaments[0].ID)
	}

	gotLive := client.GetCFLLiveGame("123")
	if gotLive.Data.BetGeniusFixtureID != "123" {
		t.Fatalf("expected live fixture ID '123', got '%s'", gotLive.Data.BetGeniusFixtureID)
	}
}
