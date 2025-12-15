package iihf

import (
	iihfClients "goalfeed/clients/leagues/iihf"
	"goalfeed/models"
	"testing"
)

func TestGetUpcomingGames_IIHF(t *testing.T) {
	svc := IIHFService{Client: iihfClients.MockIIHFApiClient{}}
	ch := make(chan []models.Game)
	go svc.GetUpcomingGames(ch)
	up := <-ch
	if len(up) == 0 {
		t.Fatalf("expected upcoming games")
	}
}
