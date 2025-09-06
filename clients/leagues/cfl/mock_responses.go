package cfl

var MockScheduleResponse = CFLScheduleResponse{
	{
		ID:        1317793,
		Status:    "complete",
		Name:      "Preseason Week 1",
		Type:      "PRE",
		Number:    1,
		StartDate: "2025-05-19T00:00:00+00:00",
		EndDate:   "2025-05-20T23:59:00+00:00",
		Tournaments: []CFLGame{
			{
				ID:     12196500,
				Date:   "2025-05-19T20:00:00+00:00",
				Status: "complete",
				HomeSquad: CFLTeam{
					ID:        93775,
					Name:      "BC Lions",
					ShortName: "BC",
					Score:     16,
				},
				AwaySquad: CFLTeam{
					ID:        112939,
					Name:      "Calgary Stampeders",
					ShortName: "CGY",
					Score:     26,
				},
				ActivePeriod: nil,
				Timeouts: CFLTimeouts{
					Away: 1,
					Home: 1,
				},
				Possession: "None",
				CFLID:      6487,
				Clock:      "00:00",
				Winner:     112939,
				IsHidden:   false,
			},
		},
	},
}

var MockLiveGameResponse = CFLLiveGameResponse{
	Data: CFLLiveGameData{
		BetGeniusFixtureID: "11824095",
		ScoreboardInfo: CFLScoreboardInfo{
			MatchStatus:          "PreMatch",
			CurrentPhase:         "P",
			AwayScore:            0,
			HomeScore:            0,
			AwayTimeoutsLeft:     2,
			HomeTimeoutsLeft:     2,
			TotalTimeouts:        2,
			TimeRemainingInPhase: "59:59",
			Possession:           "Home",
			TotalPhases:          0,
			PhaseQualifier:       "Regular",
			ClockUnreliable:      false,
		},
		MatchInfo: CFLMatchInfo{
			RoundID:            "1279423",
			RoundName:          "Week 14",
			ScheduledStartTime: "2025-09-05T23:30:00+00:00",
			VenueName:          "TD Place Stadium",
			SeasonID:           "153593",
			SeasonName:         "2025 CFL",
			HomeTeam: CFLDetailedTeam{
				FullName:     "Ottawa RedBlacks",
				CompetitorID: "88019",
				Details: CFLTeamDetails{
					Key:          "88019",
					FirstName:    "Ottawa",
					ShortName:    "REDBLACKS",
					SecondName:   "REDBLACKS",
					Abbreviation: "OTT",
					OfficialName: "Ottawa REDBLACKS",
				},
			},
			AwayTeam: CFLDetailedTeam{
				FullName:     "BC Lions",
				CompetitorID: "93775",
				Details: CFLTeamDetails{
					Key:          "93775",
					FirstName:    "BC",
					ShortName:    "Lions",
					SecondName:   "Lions",
					Abbreviation: "BC",
					OfficialName: "BC Lions",
				},
			},
			PlayedPhases: []string{"Q1", "HL"},
		},
	},
	Sport:         "AmericanFootball",
	SportID:       17,
	CompetitionID: 1035,
}
