package webApi

import (
	"goalfeed/models"
	"testing"
)

func TestNormalizeGamesData_ActiveStatusPreserved(t *testing.T) {
	games := []models.Game{
		{
			CurrentState: models.GameState{
				Status: models.StatusUpcoming,
				Period: 1,
				Clock:  "10:00",
			},
		},
		{
			CurrentState: models.GameState{
				Status: models.StatusEnded,
			},
		},
	}

	out := normalizeGamesData(games)
	if out[0].CurrentState.Status != models.StatusActive {
		t.Fatalf("expected first game to be forced to active, got %v", out[0].CurrentState.Status)
	}
	if out[1].CurrentState.Status != models.StatusEnded {
		t.Fatalf("expected ended game to remain ended")
	}
}
