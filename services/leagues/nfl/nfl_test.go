package nfl

import (
	"goalfeed/clients/leagues/nfl"
	"goalfeed/models"
	"testing"
)

func TestNFLService_GetLeagueName(t *testing.T) {
	service := NFLService{Client: nfl.NFLMockClient{}}
	if service.GetLeagueName() != "NFL" {
		t.Errorf("Expected league name to be 'NFL', got '%s'", service.GetLeagueName())
	}
}

func TestNFLService_GetActiveGames(t *testing.T) {
	service := NFLService{Client: nfl.NFLMockClient{}}
	gamesChan := make(chan []models.Game)
	go service.GetActiveGames(gamesChan)
	games := <-gamesChan

	if len(games) == 0 {
		t.Error("Expected at least one active game")
	}

	game := games[0]
	if game.LeagueId != models.LeagueIdNFL {
		t.Errorf("Expected league ID to be %d, got %d", models.LeagueIdNFL, game.LeagueId)
	}
}

func TestNFLService_GetUpcomingGames(t *testing.T) {
	service := NFLService{Client: nfl.NFLMockClient{}}
	gamesChan := make(chan []models.Game)
	go service.GetUpcomingGames(gamesChan)
	games := <-gamesChan

	if len(games) == 0 {
		t.Error("Expected at least one upcoming game")
	}

	game := games[0]
	if game.LeagueId != models.LeagueIdNFL {
		t.Errorf("Expected league ID to be %d, got %d", models.LeagueIdNFL, game.LeagueId)
	}
}
