package nfl

import (
	"goalfeed/clients/leagues/nfl"
	"goalfeed/models"
	"goalfeed/utils"
	"strconv"
	"time"
)

const (
	STATUS_UPCOMING = "pre"
	STATUS_ACTIVE   = "in"
	STATUS_FINAL    = "post"
)

type NFLService struct {
	Client nfl.INFLAPIClient
}

var logger = utils.GetLogger()

func (s NFLService) GetLeagueName() string {
	return "NFL"
}

func (s NFLService) getSchedule() nfl.NFLScheduleResponse {
	return s.Client.GetNFLSchedule()
}

func (s NFLService) GetActiveGames(ret chan []models.Game) {
	schedule := s.getSchedule()
	var activeGames []models.Game

	for _, event := range schedule.Events {
		if gameStatusFromEvent(event) == models.StatusActive {
			activeGames = append(activeGames, s.gameFromEvent(event))
		}
	}
	ret <- activeGames
}

func (s NFLService) GetUpcomingGames(ret chan []models.Game) {
	schedule := s.getSchedule()
	var upcomingGames []models.Game

	for _, event := range schedule.Events {
		if gameStatusFromEvent(event) == models.StatusUpcoming {
			upcomingGames = append(upcomingGames, s.gameFromEvent(event))
		}
	}
	ret <- upcomingGames
}

func (s NFLService) GetGameUpdate(game models.Game, ret chan models.GameUpdate) {
	s.getGameUpdateFromScoreboard(game, ret)
}

func (s NFLService) getGameUpdateFromScoreboard(game models.Game, ret chan models.GameUpdate) {
	scoreboard := s.Client.GetNFLScoreBoard(game.GameCode)

	// Extract game info from scoreboard
	var newState models.GameState
	if len(scoreboard.Events) > 0 {
		event := scoreboard.Events[0]
		if len(event.Competitions) > 0 {
			competition := event.Competitions[0]
			if len(competition.Competitors) >= 2 {
				var awayTeam, homeTeam nfl.NFLScoreboardCompetitor

				// Find home and away teams by checking HomeAway field
				for _, competitor := range competition.Competitors {
					if competitor.HomeAway == "home" {
						homeTeam = competitor
					} else if competitor.HomeAway == "away" {
						awayTeam = competitor
					}
				}

				awayScore, _ := strconv.Atoi(awayTeam.Score)
				homeScore, _ := strconv.Atoi(homeTeam.Score)

				newState = models.GameState{
					Home: models.TeamState{
						Team:  s.teamFromCompetitor(homeTeam),
						Score: homeScore,
					},
					Away: models.TeamState{
						Team:  s.teamFromCompetitor(awayTeam),
						Score: awayScore,
					},
					Status:     gameStatusFromEventStatus(event.Status),
					Period:     event.Status.Period,
					PeriodType: "QUARTER",
					Clock:      event.Status.DisplayClock,
					Venue: models.Venue{
						Name: competition.Venue.FullName,
					},
				}
			}
		}
	}

	ret <- models.GameUpdate{
		OldState: game.CurrentState,
		NewState: newState,
	}
}

func (s NFLService) teamFromCompetitor(competitor nfl.NFLScoreboardCompetitor) models.Team {
	return models.Team{
		TeamName: competitor.Team.DisplayName,
		TeamCode: competitor.Team.Abbreviation,
		ExtID:    competitor.Team.ID,
		LeagueID: models.LeagueIdNFL,
		LogoURL:  competitor.Team.Logo,
	}
}

func (s NFLService) gameFromEvent(event nfl.NFLScheduleEvent) models.Game {
	var homeTeam, awayTeam nfl.NFLCompetitor
	var venue models.Venue
	var gameDate time.Time

	if len(event.Competitions) > 0 {
		competition := event.Competitions[0]
		venue = models.Venue{
			Name:  competition.Venue.FullName,
			City:  competition.Venue.Address.City,
			State: competition.Venue.Address.State,
		}

		if len(competition.Competitors) >= 2 {
			// Find home and away teams by checking HomeAway field
			for _, competitor := range competition.Competitors {
				if competitor.HomeAway == "home" {
					homeTeam = competitor
				} else if competitor.HomeAway == "away" {
					awayTeam = competitor
				}
			}
		}
	}

	// Parse the game date with multiple layouts
	if event.Date != "" {
		layouts := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05-07:00",
			"2006-01-02T15:04Z",
			"2006-01-02T15:04-07:00",
		}
		for _, layout := range layouts {
			if parsedDate, err := time.Parse(layout, event.Date); err == nil {
				gameDate = parsedDate
				break
			}
		}
	}

	// Fallback to competition date if event.Date is empty or failed parsing
	if gameDate.IsZero() && len(event.Competitions) > 0 {
		competition := event.Competitions[0]
		if competition.Date != "" {
			layouts := []string{
				time.RFC3339,
				time.RFC3339Nano,
				"2006-01-02T15:04:05Z",
				"2006-01-02T15:04:05-07:00",
				"2006-01-02T15:04Z",
				"2006-01-02T15:04-07:00",
			}
			for _, layout := range layouts {
				if parsedDate, err := time.Parse(layout, competition.Date); err == nil {
					gameDate = parsedDate
					break
				}
			}
		}
	}

	// Final fallback: fetch scoreboard summary to get reliable date
	if gameDate.IsZero() && event.ID != "" {
		summary := s.Client.GetNFLScoreBoard(event.ID)
		if len(summary.Events) > 0 {
			se := summary.Events[0]
			// Try event date first
			layouts := []string{
				time.RFC3339,
				time.RFC3339Nano,
				"2006-01-02T15:04:05Z",
				"2006-01-02T15:04:05-07:00",
				"2006-01-02T15:04Z",
				"2006-01-02T15:04-07:00",
			}
			if se.Date != "" {
				for _, layout := range layouts {
					if parsedDate, err := time.Parse(layout, se.Date); err == nil {
						gameDate = parsedDate
						break
					}
				}
			}
			// Then try competition date from summary
			if gameDate.IsZero() && len(se.Competitions) > 0 {
				cd := se.Competitions[0].Date
				if cd != "" {
					for _, layout := range layouts {
						if parsedDate, err := time.Parse(layout, cd); err == nil {
							gameDate = parsedDate
							break
						}
					}
				}
			}
		}
	}

	awayScore, _ := strconv.Atoi(awayTeam.Score)
	homeScore, _ := strconv.Atoi(homeTeam.Score)

	// Format game time for upcoming games
	var gameTimeDisplay string
	if event.Status.Type.State == STATUS_UPCOMING {
		if !gameDate.IsZero() {
			gameTimeDisplay = gameDate.Format("Mon 3:04 PM")
		} else {
			gameTimeDisplay = "TBD"
		}
	} else {
		gameTimeDisplay = event.Status.DisplayClock
	}

	// Precompute display time for details
	gameTimeStr := "TBD"
	if !gameDate.IsZero() {
		gameTimeStr = gameDate.Format("3:04 PM")
	}

	return models.Game{
		CurrentState: models.GameState{
			ExtTimestamp: event.Date,
			Home: models.TeamState{
				Team:  s.teamFromScheduleCompetitor(homeTeam),
				Score: homeScore,
			},
			Away: models.TeamState{
				Team:  s.teamFromScheduleCompetitor(awayTeam),
				Score: awayScore,
			},
			Status:     gameStatusFromEvent(event),
			FetchedAt:  time.Now(),
			Period:     event.Status.Period,
			PeriodType: "QUARTER",
			Clock:      gameTimeDisplay,
			Venue:      venue,
		},
		GameCode:     event.ID,
		ExtTimestamp: event.Date,
		LeagueId:     models.LeagueIdNFL,
		GameDetails: models.GameDetails{
			GameId:     event.ID,
			Season:     strconv.Itoa(event.Season.Year),
			SeasonType: "REGULAR",
			Week:       event.Week.Number,
			GameDate:   gameDate,
			GameTime:   gameTimeStr,
			Timezone:   "UTC", // TODO: Get actual timezone from venue
		},
	}
}

func (s NFLService) teamFromScheduleCompetitor(competitor nfl.NFLCompetitor) models.Team {
	return models.Team{
		TeamName: competitor.Team.DisplayName,
		TeamCode: competitor.Team.Abbreviation,
		ExtID:    competitor.Team.ID,
		LeagueID: models.LeagueIdNFL,
		LogoURL:  competitor.Team.Logo,
	}
}

func gameStatusFromEvent(event nfl.NFLScheduleEvent) models.GameStatus {
	return gameStatusFromEventStatus(event.Status)
}

func gameStatusFromEventStatus(status struct {
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
}) models.GameStatus {
	switch status.Type.State {
	case STATUS_FINAL:
		return models.StatusEnded
	case STATUS_UPCOMING:
		return models.StatusUpcoming
	case STATUS_ACTIVE:
		return models.StatusActive
	default:
		return models.StatusUpcoming
	}
}

func (s NFLService) GetEvents(update models.GameUpdate, ret chan []models.Event) {
	events := append(
		s.getTouchdownEvents(update.OldState.Home, update.NewState.Home, update.OldState.Away.Team),
		s.getTouchdownEvents(update.OldState.Away, update.NewState.Away, update.OldState.Home.Team)...,
	)
	ret <- events
}

func (s NFLService) getTouchdownEvents(oldState models.TeamState, newState models.TeamState, opponent models.Team) []models.Event {
	events := []models.Event{}
	diff := newState.Score - oldState.Score
	team := newState.Team

	// In NFL, we track touchdowns (7 points) and field goals (3 points)
	// For simplicity, we'll create events for each point scored
	for i := 0; i < diff; i++ {
		events = append(events, models.Event{
			TeamCode:     team.TeamCode,
			TeamName:     team.TeamName,
			TeamHash:     team.GetTeamHash(),
			LeagueId:     models.LeagueIdNFL,
			LeagueName:   s.GetLeagueName(),
			OpponentCode: opponent.TeamCode,
			OpponentName: opponent.TeamName,
			OpponentHash: opponent.GetTeamHash(),
		})
	}
	return events
}
