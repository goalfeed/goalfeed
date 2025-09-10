package cfl

import (
	"goalfeed/clients/leagues/cfl"
	"goalfeed/models"
	"testing"
)

func TestCFLService_GetLeagueName(t *testing.T) {
	service := CFLService{}
	if service.GetLeagueName() != "CFL" {
		t.Errorf("Expected league name to be 'CFL', got '%s'", service.GetLeagueName())
	}
}

func TestCFLService_GetActiveGames(t *testing.T) {
	mockClient := cfl.MockCFLApiClient{
		ScheduleResponse: cfl.MockScheduleResponse,
	}
	
	service := CFLService{Client: mockClient}
	
	ret := make(chan []models.Game)
	go service.GetActiveGames(ret)
	
	games := <-ret
	
	// Should have 0 active games since the mock data has completed games
	if len(games) != 0 {
		t.Errorf("Expected 0 active games, got %d", len(games))
	}
}

func TestCFLService_GetGameUpdate(t *testing.T) {
	mockClient := cfl.MockCFLApiClient{
		LiveGameResponse: cfl.MockLiveGameResponse,
	}
	
	service := CFLService{Client: mockClient}
	
	// Create a mock game
	game := models.Game{
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team: models.Team{
					TeamName: "Ottawa RedBlacks",
					TeamCode: "OTT",
					LeagueID: models.LeagueIdCFL,
				},
				Score: 0,
			},
			Away: models.TeamState{
				Team: models.Team{
					TeamName: "BC Lions",
					TeamCode: "BC",
					LeagueID: models.LeagueIdCFL,
				},
				Score: 0,
			},
			Status: models.StatusUpcoming,
		},
		GameCode: "11824095",
		LeagueId: models.LeagueIdCFL,
	}
	
	ret := make(chan models.GameUpdate)
	go service.GetGameUpdate(game, ret)
	
	update := <-ret
	
	if update.NewState.Home.Score != 0 {
		t.Errorf("Expected home score to be 0, got %d", update.NewState.Home.Score)
	}
	
	if update.NewState.Away.Score != 0 {
		t.Errorf("Expected away score to be 0, got %d", update.NewState.Away.Score)
	}
}

func TestCFLService_GetEvents(t *testing.T) {
	service := CFLService{}
	
	// Create a mock game update with score change
	update := models.GameUpdate{
		OldState: models.GameState{
			Home: models.TeamState{
				Team: models.Team{
					TeamName: "Ottawa RedBlacks",
					TeamCode: "OTT",
					LeagueID: models.LeagueIdCFL,
				},
				Score: 0,
			},
			Away: models.TeamState{
				Team: models.Team{
					TeamName: "BC Lions",
					TeamCode: "BC",
					LeagueID: models.LeagueIdCFL,
				},
				Score: 0,
			},
		},
		NewState: models.GameState{
			Home: models.TeamState{
				Team: models.Team{
					TeamName: "Ottawa RedBlacks",
					TeamCode: "OTT",
					LeagueID: models.LeagueIdCFL,
				},
				Score: 6, // Touchdown scored
			},
			Away: models.TeamState{
				Team: models.Team{
					TeamName: "BC Lions",
					TeamCode: "BC",
					LeagueID: models.LeagueIdCFL,
				},
				Score: 0,
			},
		},
	}
	
	ret := make(chan []models.Event)
	go service.GetEvents(update, ret)
	
	events := <-ret
	
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
	
	if events[0].TeamCode != "OTT" {
		t.Errorf("Expected team code to be 'OTT', got '%s'", events[0].TeamCode)
	}
}
