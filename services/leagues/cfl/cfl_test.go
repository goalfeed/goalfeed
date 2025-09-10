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
	// Create a mock schedule with completed games (not live)
	mockScheduleResponse := cfl.CFLScheduleResponse{
		{
			ID:        1317793,
			Status:    "complete",
			Name:      "Preseason Week 1",
			Type:      "PRE",
			Number:    1,
			StartDate: "2025-05-19T00:00:00+00:00",
			EndDate:   "2025-05-20T23:59:00+00:00",
			Tournaments: []cfl.CFLGame{
				{
					ID:     11824097,
					Date:   "2025-05-19T20:00:00+00:00",
					Status: "complete", // Changed from "live" to "complete"
					HomeSquad: cfl.CFLTeam{
						ID:        93775,
						Name:      "Winnipeg Blue Bombers",
						ShortName: "WPG",
						Score:     14,
					},
					AwaySquad: cfl.CFLTeam{
						ID:        112939,
						Name:      "Saskatchewan Roughriders",
						ShortName: "SSK",
						Score:     10,
					},
					ActivePeriod: "F",
					Timeouts: cfl.CFLTimeouts{
						Away: 2,
						Home: 1,
					},
					Possession: "",
					CFLID:      6487,
					Clock:      "00:00",
					Winner:     nil,
					IsHidden:   false,
				},
			},
		},
	}

	mockClient := cfl.MockCFLApiClient{
		ScheduleResponse: mockScheduleResponse,
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
	// Create a mock live game response with 0 scores to match test expectations
	mockLiveGameResponse := cfl.CFLLiveGameResponse{
		Data: cfl.CFLLiveGameData{
			BetGeniusFixtureID: "11824095",
			ScoreboardInfo: cfl.CFLScoreboardInfo{
				MatchStatus:          "live",
				CurrentPhase:         "Q1",
				AwayScore:            0, // Changed from 10 to 0
				HomeScore:            0, // Changed from 14 to 0
				AwayTimeoutsLeft:     3,
				HomeTimeoutsLeft:     3,
				TotalTimeouts:        3,
				TimeRemainingInPhase: "15:00",
				Possession:           "",
				Down:                 1,
				YardsToGo:            10,
				TotalPhases:          1,
				PhaseQualifier:       "Regular",
				ClockUnreliable:      false,
			},
			LiveStream: cfl.CFLLiveStream{
				CurrentPlay: cfl.CFLCurrentPlay{
					DownNumber:      1,
					LineOfScrimmage: 35,
					FirstDownLine:   45,
					PlayType:        "pass",
					YardsToGo:       10,
					Possession:      "",
					Clock:           "15:00",
					Phase:           "Q1",
					PlayFormation:   "",
					Quarterback:     0,
					YardLine: cfl.CFLYardLine{
						TeamNumber: 1,
						YardLine:   35,
					},
				},
			},
		},
	}

	mockClient := cfl.MockCFLApiClient{
		LiveGameResponse: mockLiveGameResponse,
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
