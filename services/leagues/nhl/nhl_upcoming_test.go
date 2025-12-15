package nhl

import (
	nhlClients "goalfeed/clients/leagues/nhl"
	"goalfeed/models"
	"testing"
)

func TestGetUpcomingGames_NHL(t *testing.T) {
	svc := NHLService{Client: nhlClients.MockNHLApiClient{}}
	ch := make(chan []models.Game)
	go svc.GetUpcomingGames(ch)
	up := <-ch
	if len(up) == 0 {
		t.Fatalf("expected upcoming games")
	}
}
