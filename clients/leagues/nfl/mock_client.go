package nfl

import "time"

type NFLMockClient struct {
}

func (c NFLMockClient) GetNFLSchedule() NFLScheduleResponse {
	return NFLScheduleResponse{
		Events: []NFLScheduleEvent{
			{
				ID:        "401547403",
				UID:       "s:20~l:28~e:401547403",
				Date:      time.Now().Add(2 * time.Hour).Format(time.RFC3339),
				Name:      "Buffalo Bills vs. Miami Dolphins",
				ShortName: "BUF @ MIA",
				Season: struct {
					Year int `json:"year"`
				}{
					Year: 2024,
				},
				Week: struct {
					Number int `json:"number"`
				}{
					Number: 1,
				},
				Competitions: []NFLCompetition{
					{
						ID:   "401547403",
						UID:  "s:20~l:28~e:401547403~c:401547403",
						Date: time.Now().Add(2 * time.Hour).Format("2006-01-02T15:04:05Z"),
						Venue: struct {
							ID       string `json:"id"`
							FullName string `json:"fullName"`
							Address  struct {
								City  string `json:"city"`
								State string `json:"state"`
							} `json:"address"`
							Grass  bool `json:"grass"`
							Indoor bool `json:"indoor"`
						}{
							ID:       "1",
							FullName: "Hard Rock Stadium",
							Address: struct {
								City  string `json:"city"`
								State string `json:"state"`
							}{
								City:  "Miami Gardens",
								State: "FL",
							},
							Grass:  true,
							Indoor: false,
						},
						Competitors: []NFLCompetitor{
							{
								ID:       "401547403-1",
								UID:      "s:20~l:28~e:401547403~c:401547403~t:1",
								Type:     "team",
								Order:    0,
								HomeAway: "away",
								Winner:   false,
								Team: struct {
									ID               string `json:"id"`
									UID              string `json:"uid"`
									Location         string `json:"location"`
									Name             string `json:"name"`
									Abbreviation     string `json:"abbreviation"`
									DisplayName      string `json:"displayName"`
									ShortDisplayName string `json:"shortDisplayName"`
									Color            string `json:"color"`
									AlternateColor   string `json:"alternateColor"`
									IsActive         bool   `json:"isActive"`
									Venue            struct {
										ID string `json:"id"`
									} `json:"venue"`
									Links []struct {
										Rel  []string `json:"rel"`
										Href string   `json:"href"`
										Text string   `json:"text"`
									} `json:"links"`
									Logo string `json:"logo"`
								}{
									ID:               "4",
									UID:              "s:20~l:28~t:4",
									Location:         "Buffalo",
									Name:             "Bills",
									Abbreviation:     "BUF",
									DisplayName:      "Buffalo Bills",
									ShortDisplayName: "Bills",
									Color:            "00338D",
									AlternateColor:   "C60C30",
									IsActive:         true,
									Logo:             "https://a.espncdn.com/i/teamlogos/nfl/500/buf.png",
								},
								Score: "0",
							},
							{
								ID:       "401547403-2",
								UID:      "s:20~l:28~e:401547403~c:401547403~t:2",
								Type:     "team",
								Order:    1,
								HomeAway: "home",
								Winner:   false,
								Team: struct {
									ID               string `json:"id"`
									UID              string `json:"uid"`
									Location         string `json:"location"`
									Name             string `json:"name"`
									Abbreviation     string `json:"abbreviation"`
									DisplayName      string `json:"displayName"`
									ShortDisplayName string `json:"shortDisplayName"`
									Color            string `json:"color"`
									AlternateColor   string `json:"alternateColor"`
									IsActive         bool   `json:"isActive"`
									Venue            struct {
										ID string `json:"id"`
									} `json:"venue"`
									Links []struct {
										Rel  []string `json:"rel"`
										Href string   `json:"href"`
										Text string   `json:"text"`
									} `json:"links"`
									Logo string `json:"logo"`
								}{
									ID:               "15",
									UID:              "s:20~l:28~t:15",
									Location:         "Miami",
									Name:             "Dolphins",
									Abbreviation:     "MIA",
									DisplayName:      "Miami Dolphins",
									ShortDisplayName: "Dolphins",
									Color:            "008E97",
									AlternateColor:   "FC4C02",
									IsActive:         true,
									Logo:             "https://a.espncdn.com/i/teamlogos/nfl/500/mia.png",
								},
								Score: "0",
							},
						},
					},
				},
				Status: struct {
					Clock        float64 `json:"clock"`
					DisplayClock string  `json:"displayClock"`
					Period       int     `json:"period"`
					Type         struct {
						ID          string `json:"id"`
						Name        string `json:"name"`
						State       string `json:"state"`
						Completed   bool   `json:"completed"`
						Description string `json:"description"`
						Detail      string `json:"detail"`
						ShortDetail string `json:"shortDetail"`
					} `json:"type"`
				}{
					Clock:        1200,
					DisplayClock: "20:00",
					Period:       2,
					Type: struct {
						ID          string `json:"id"`
						Name        string `json:"name"`
						State       string `json:"state"`
						Completed   bool   `json:"completed"`
						Description string `json:"description"`
						Detail      string `json:"detail"`
						ShortDetail string `json:"shortDetail"`
					}{
						ID:          "3",
						Name:        "STATUS_IN_PROGRESS",
						State:       "in",
						Completed:   false,
						Description: "In Progress",
						Detail:      "2nd Quarter",
						ShortDetail: "2Q",
					},
				},
			},
			{
				ID:        "401547404",
				UID:       "s:20~l:28~e:401547404",
				Date:      time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				Name:      "Kansas City Chiefs vs. Dallas Cowboys",
				ShortName: "KC @ DAL",
				Season: struct {
					Year int `json:"year"`
				}{
					Year: 2024,
				},
				Week: struct {
					Number int `json:"number"`
				}{
					Number: 1,
				},
				Competitions: []NFLCompetition{
					{
						ID:   "401547404",
						UID:  "s:20~l:28~e:401547404~c:401547404",
						Date: time.Now().Add(24 * time.Hour).Format("2006-01-02T15:04:05Z"),
						Venue: struct {
							ID       string `json:"id"`
							FullName string `json:"fullName"`
							Address  struct {
								City  string `json:"city"`
								State string `json:"state"`
							} `json:"address"`
							Grass  bool `json:"grass"`
							Indoor bool `json:"indoor"`
						}{
							ID:       "2",
							FullName: "AT&T Stadium",
							Address: struct {
								City  string `json:"city"`
								State string `json:"state"`
							}{
								City:  "Arlington",
								State: "TX",
							},
							Grass:  true,
							Indoor: false,
						},
						Competitors: []NFLCompetitor{
							{
								ID:       "401547404-1",
								UID:      "s:20~l:28~e:401547404~c:401547404~t:1",
								Type:     "team",
								Order:    0,
								HomeAway: "away",
								Winner:   false,
								Team: struct {
									ID               string `json:"id"`
									UID              string `json:"uid"`
									Location         string `json:"location"`
									Name             string `json:"name"`
									Abbreviation     string `json:"abbreviation"`
									DisplayName      string `json:"displayName"`
									ShortDisplayName string `json:"shortDisplayName"`
									Color            string `json:"color"`
									AlternateColor   string `json:"alternateColor"`
									IsActive         bool   `json:"isActive"`
									Venue            struct {
										ID string `json:"id"`
									} `json:"venue"`
									Links []struct {
										Rel  []string `json:"rel"`
										Href string   `json:"href"`
										Text string   `json:"text"`
									} `json:"links"`
									Logo string `json:"logo"`
								}{
									ID:               "12",
									UID:              "s:20~l:28~t:12",
									Location:         "Kansas City",
									Name:             "Chiefs",
									Abbreviation:     "KC",
									DisplayName:      "Kansas City Chiefs",
									ShortDisplayName: "Chiefs",
									Color:            "E31837",
									AlternateColor:   "FFB81C",
									IsActive:         true,
									Logo:             "https://a.espncdn.com/i/teamlogos/nfl/500/kc.png",
								},
								Score: "0",
							},
							{
								ID:       "401547404-2",
								UID:      "s:20~l:28~e:401547404~c:401547404~t:2",
								Type:     "team",
								Order:    1,
								HomeAway: "home",
								Winner:   false,
								Team: struct {
									ID               string `json:"id"`
									UID              string `json:"uid"`
									Location         string `json:"location"`
									Name             string `json:"name"`
									Abbreviation     string `json:"abbreviation"`
									DisplayName      string `json:"displayName"`
									ShortDisplayName string `json:"shortDisplayName"`
									Color            string `json:"color"`
									AlternateColor   string `json:"alternateColor"`
									IsActive         bool   `json:"isActive"`
									Venue            struct {
										ID string `json:"id"`
									} `json:"venue"`
									Links []struct {
										Rel  []string `json:"rel"`
										Href string   `json:"href"`
										Text string   `json:"text"`
									} `json:"links"`
									Logo string `json:"logo"`
								}{
									ID:               "6",
									UID:              "s:20~l:28~t:6",
									Location:         "Dallas",
									Name:             "Cowboys",
									Abbreviation:     "DAL",
									DisplayName:      "Dallas Cowboys",
									ShortDisplayName: "Cowboys",
									Color:            "003594",
									AlternateColor:   "869397",
									IsActive:         true,
									Logo:             "https://a.espncdn.com/i/teamlogos/nfl/500/dal.png",
								},
								Score: "0",
							},
						},
					},
				},
				Status: struct {
					Clock        float64 `json:"clock"`
					DisplayClock string  `json:"displayClock"`
					Period       int     `json:"period"`
					Type         struct {
						ID          string `json:"id"`
						Name        string `json:"name"`
						State       string `json:"state"`
						Completed   bool   `json:"completed"`
						Description string `json:"description"`
						Detail      string `json:"detail"`
						ShortDetail string `json:"shortDetail"`
					} `json:"type"`
				}{
					Clock:        0,
					DisplayClock: "0:00",
					Period:       0,
					Type: struct {
						ID          string `json:"id"`
						Name        string `json:"name"`
						State       string `json:"state"`
						Completed   bool   `json:"completed"`
						Description string `json:"description"`
						Detail      string `json:"detail"`
						ShortDetail string `json:"shortDetail"`
					}{
						ID:          "1",
						Name:        "STATUS_SCHEDULED",
						State:       "pre",
						Completed:   false,
						Description: "Scheduled",
						Detail:      "Mon, Sep 9 at 8:15 PM ET",
						ShortDetail: "9/9 - 8:15 PM ET",
					},
				},
			},
		},
	}
}

func (c NFLMockClient) GetNFLScoreBoard(gameId string) NFLScoreboardResponse {
	return NFLScoreboardResponse{
		Events: []NFLScoreboardEvent{
			{
				ID:        gameId,
				UID:       "s:20~l:28~e:" + gameId,
				Date:      time.Now().Format("2006-01-02T15:04:05Z"),
				Name:      "Buffalo Bills vs. Miami Dolphins",
				ShortName: "BUF @ MIA",
				Competitions: []NFLScoreboardCompetition{
					{
						ID: gameId,
						Competitors: []NFLScoreboardCompetitor{
							{
								ID:       gameId + "-1",
								HomeAway: "away",
								Team: struct {
									ID               string `json:"id"`
									UID              string `json:"uid"`
									Location         string `json:"location"`
									Name             string `json:"name"`
									Abbreviation     string `json:"abbreviation"`
									DisplayName      string `json:"displayName"`
									ShortDisplayName string `json:"shortDisplayName"`
									Color            string `json:"color"`
									AlternateColor   string `json:"alternateColor"`
									IsActive         bool   `json:"isActive"`
									Venue            struct {
										ID string `json:"id"`
									} `json:"venue"`
									Links []struct {
										Rel  []string `json:"rel"`
										Href string   `json:"href"`
										Text string   `json:"text"`
									} `json:"links"`
									Logo string `json:"logo"`
								}{
									ID:               "4",
									Location:         "Buffalo",
									Name:             "Bills",
									Abbreviation:     "BUF",
									DisplayName:      "Buffalo Bills",
									ShortDisplayName: "Bills",
									Color:            "00338D",
									IsActive:         true,
									Logo:             "https://a.espncdn.com/i/teamlogos/nfl/500/buf.png",
								},
								Score: "21",
							},
							{
								ID:       gameId + "-2",
								HomeAway: "home",
								Team: struct {
									ID               string `json:"id"`
									UID              string `json:"uid"`
									Location         string `json:"location"`
									Name             string `json:"name"`
									Abbreviation     string `json:"abbreviation"`
									DisplayName      string `json:"displayName"`
									ShortDisplayName string `json:"shortDisplayName"`
									Color            string `json:"color"`
									AlternateColor   string `json:"alternateColor"`
									IsActive         bool   `json:"isActive"`
									Venue            struct {
										ID string `json:"id"`
									} `json:"venue"`
									Links []struct {
										Rel  []string `json:"rel"`
										Href string   `json:"href"`
										Text string   `json:"text"`
									} `json:"links"`
									Logo string `json:"logo"`
								}{
									ID:               "15",
									Location:         "Miami",
									Name:             "Dolphins",
									Abbreviation:     "MIA",
									DisplayName:      "Miami Dolphins",
									ShortDisplayName: "Dolphins",
									Color:            "008E97",
									IsActive:         true,
									Logo:             "https://a.espncdn.com/i/teamlogos/nfl/500/mia.png",
								},
								Score: "17",
							},
						},
					},
				},
				Status: struct {
					Clock        float64 `json:"clock"`
					DisplayClock string  `json:"displayClock"`
					Period       int     `json:"period"`
					Type         struct {
						ID          string `json:"id"`
						Name        string `json:"name"`
						State       string `json:"state"`
						Completed   bool   `json:"completed"`
						Description string `json:"description"`
						Detail      string `json:"detail"`
						ShortDetail string `json:"shortDetail"`
					} `json:"type"`
				}{
					Clock:        1200,
					DisplayClock: "20:00",
					Period:       4,
					Type: struct {
						ID          string `json:"id"`
						Name        string `json:"name"`
						State       string `json:"state"`
						Completed   bool   `json:"completed"`
						Description string `json:"description"`
						Detail      string `json:"detail"`
						ShortDetail string `json:"shortDetail"`
					}{
						ID:          "3",
						Name:        "STATUS_IN_PROGRESS",
						State:       "in",
						Completed:   false,
						Description: "In Progress",
						Detail:      "4th Quarter",
						ShortDetail: "4Q",
					},
				},
			},
		},
	}
}

func (c NFLMockClient) GetTeam(teamAbbr string) NFLTeamResponse {
	return NFLTeamResponse{
		Teams: []NFLTeam{
			{
				ID:               "4",
				UID:              "s:20~l:28~t:4",
				Location:         "Buffalo",
				Name:             "Bills",
				Abbreviation:     "BUF",
				DisplayName:      "Buffalo Bills",
				ShortDisplayName: "Bills",
				Color:            "00338D",
				AlternateColor:   "C60C30",
				IsActive:         true,
				Logo:             "https://a.espncdn.com/i/teamlogos/nfl/500/buf.png",
			},
		},
	}
}

func (c NFLMockClient) GetAllTeams() NFLTeamResponse {
	return NFLTeamResponse{
		Teams: []NFLTeam{
			{
				ID:               "4",
				UID:              "s:20~l:28~t:4",
				Location:         "Buffalo",
				Name:             "Bills",
				Abbreviation:     "BUF",
				DisplayName:      "Buffalo Bills",
				ShortDisplayName: "Bills",
				Color:            "00338D",
				AlternateColor:   "C60C30",
				IsActive:         true,
				Logo:             "https://a.espncdn.com/i/teamlogos/nfl/500/buf.png",
			},
		},
	}
}
