package nhl

import (
	"fmt"
	"goalfeed/clients/leagues/nhl"
	"goalfeed/models"
	"goalfeed/utils"
	"strconv"
	"strings"
	"time"
)

const (
	STATUS_UPCOMING = "PRE"
	STATUS_OFF      = "OFF"
	STATUS_FUT      = "FUT"
	STATUS_ACTIVE   = "LIVE"
	STATUS_FINAL    = "FINAL"
)

type NHLService struct {
	Client nhl.INHLApiClient
}

var logger = utils.GetLogger()

func (s NHLService) GetLeagueName() string {
	return "NHL"
}

func (s NHLService) getSchedule() nhl.NHLScheduleResponse {
	return s.Client.GetNHLSchedule()
}

func (s NHLService) getScheduleByDate(date string) nhl.NHLScheduleResponse {
	return s.Client.GetNHLScheduleByDate(date)
}

func (s NHLService) GetActiveGames(ret chan []models.Game) {
	schedule := s.getSchedule()
	var activeGames []models.Game

	for _, date := range schedule.GameWeek {
		for _, game := range date.Games {
			if gameStatusFromScheduleGame(game) == models.StatusActive {
				// For live games, always fetch from scoreboard to get accurate data
				if game.GameState == "LIVE" {
					fresh := s.gameFromScoreboard(strconv.Itoa(game.ID))
					// Validate scoreboard response has valid data (check ID matches)
					if fresh.GameCode != "" && fresh.GameCode == strconv.Itoa(game.ID) {
						// Use the complete game data from scoreboard
						activeGames = append(activeGames, fresh)
					} else {
						// Fallback to schedule data if scoreboard fails
						activeGames = append(activeGames, s.gameFromSchedule(game))
					}
				} else {
					// For non-LIVE active games (e.g., FINAL), use schedule data
					activeGames = append(activeGames, s.gameFromSchedule(game))
				}
			}
		}
	}
	ret <- activeGames
}

func (s NHLService) GetUpcomingGames(ret chan []models.Game) {
	schedule := s.getSchedule()
	var upcomingGames []models.Game

	for _, date := range schedule.GameWeek {
		for _, game := range date.Games {
			if gameStatusFromScheduleGame(game) == models.StatusUpcoming {
				upcomingGames = append(upcomingGames, s.gameFromSchedule(game))
			}
		}
	}
	ret <- upcomingGames
}

func (s NHLService) GetGamesByDate(date string, ret chan []models.Game) {
	schedule := s.getScheduleByDate(date)
	var games []models.Game

	for _, dateGroup := range schedule.GameWeek {
		for _, game := range dateGroup.Games {
			// Include all games (active, ended, upcoming) for the specified date
			gameModel := s.gameFromSchedule(game)
			// For completed games, try to get final score from scoreboard if available
			if gameStatusFromScheduleGame(game) == models.StatusEnded {
				// Try to get final score from scoreboard
				fresh := s.gameFromScoreboard(strconv.Itoa(game.ID))
				if fresh.GameCode != "" {
					gameModel = fresh
				}
			}
			games = append(games, gameModel)
		}
	}
	ret <- games
}

func (s NHLService) GetGameUpdate(game models.Game, ret chan models.GameUpdate) {
	s.getGameUpdateFromScoreboard(game, ret)
}
func fudgeTimestamp(extTimestamp string) string {

	pieces := strings.Split(extTimestamp, "_")
	oldTimeInt, _ := strconv.Atoi(pieces[1])
	newTimeInt := oldTimeInt - 10
	_ = pieces
	newTime := fmt.Sprintf("%s_%06d", pieces[0], newTimeInt)
	return newTime

}

// gameFromScoreboard builds a complete Game model from the scoreboard endpoint
func (s NHLService) gameFromScoreboard(gameId string) models.Game {
	scoreboard := s.Client.GetNHLScoreBoard(gameId)

	// Extract period information from periodDescriptor
	var period int
	var periodType string
	var clock string

	if scoreboard.PeriodDescriptor.Number > 0 {
		period = scoreboard.PeriodDescriptor.Number
		// Map period type: "REG" -> "REGULAR", "OT" -> "OVERTIME", etc.
		switch scoreboard.PeriodDescriptor.PeriodType {
		case "REG":
			periodType = "REGULAR"
		case "OT":
			periodType = "OVERTIME"
		case "SO":
			periodType = "SHOOTOUT"
		default:
			periodType = "REGULAR"
		}
	} else if scoreboard.GameState == "LIVE" {
		// Fallback to period 1 if periodDescriptor is missing but game is live
		period = 1
		periodType = "REGULAR"
	}

	// Extract clock time from clock object
	if scoreboard.Clock.TimeRemaining != "" {
		clock = scoreboard.Clock.TimeRemaining
	} else if scoreboard.GameState == "LIVE" {
		clock = "LIVE"
	}

	return models.Game{
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team: models.Team{
					TeamName: scoreboard.HomeTeam.PlaceName.Default,
					TeamCode: scoreboard.HomeTeam.Abbrev,
					ExtID:    scoreboard.HomeTeam.Abbrev,
					LeagueID: models.LeagueIdNHL,
					LogoURL:  scoreboard.HomeTeam.Logo,
				},
				Score: scoreboard.HomeTeam.Score,
				Statistics: models.TeamStats{
					Shots: scoreboard.HomeTeam.Sog,
				},
			},
			Away: models.TeamState{
				Team: models.Team{
					TeamName: scoreboard.AwayTeam.PlaceName.Default,
					TeamCode: scoreboard.AwayTeam.Abbrev,
					ExtID:    scoreboard.AwayTeam.Abbrev,
					LeagueID: models.LeagueIdNHL,
					LogoURL:  scoreboard.AwayTeam.Logo,
				},
				Score: scoreboard.AwayTeam.Score,
				Statistics: models.TeamStats{
					Shots: scoreboard.AwayTeam.Sog,
				},
			},
			Status:        gameStatusFromGameState(scoreboard.GameState),
			FetchedAt:     time.Now(),
			Period:        period,
			PeriodType:    periodType,
			Clock:         clock,
			TimeRemaining: clock, // Set both for frontend compatibility
			Venue: models.Venue{
				Name: scoreboard.Venue.Default,
			},
		},
		GameCode: strconv.Itoa(scoreboard.ID),
		LeagueId: models.LeagueIdNHL,
		GameDetails: models.GameDetails{
			GameId:     strconv.Itoa(scoreboard.ID),
			Season:     strconv.Itoa(scoreboard.Season),
			SeasonType: strconv.Itoa(scoreboard.GameType),
			GameDate:   scoreboard.StartTimeUTC,
			GameTime:   scoreboard.StartTimeUTC.Format("3:04 PM"),
			Timezone:   "UTC",
		},
	}
}

func (s NHLService) getGameUpdateFromScoreboard(game models.Game, ret chan models.GameUpdate) {
	scoreboard := s.Client.GetNHLScoreBoard(game.GameCode)

	// Extract period information from periodDescriptor
	var period int
	var periodType string
	var clock string

	if scoreboard.PeriodDescriptor.Number > 0 {
		period = scoreboard.PeriodDescriptor.Number
		// Map period type: "REG" -> "REGULAR", "OT" -> "OVERTIME", etc.
		switch scoreboard.PeriodDescriptor.PeriodType {
		case "REG":
			periodType = "REGULAR"
		case "OT":
			periodType = "OVERTIME"
		case "SO":
			periodType = "SHOOTOUT"
		default:
			periodType = "REGULAR"
		}
	} else if scoreboard.GameState == "LIVE" {
		// Fallback to period 1 if periodDescriptor is missing but game is live
		period = 1
		periodType = "REGULAR"
	}

	// Extract clock time from clock object
	if scoreboard.Clock.TimeRemaining != "" {
		clock = scoreboard.Clock.TimeRemaining
	} else if scoreboard.GameState == "LIVE" {
		clock = "LIVE"
	}

	newState := models.GameState{
		Home: models.TeamState{
			Team:  game.CurrentState.Home.Team,
			Score: scoreboard.HomeTeam.Score,
			Statistics: models.TeamStats{
				Shots: scoreboard.HomeTeam.Sog,
			},
		},
		Away: models.TeamState{
			Team:  game.CurrentState.Away.Team,
			Score: scoreboard.AwayTeam.Score,
			Statistics: models.TeamStats{
				Shots: scoreboard.AwayTeam.Sog,
			},
		},
		Status:        gameStatusFromGameState(scoreboard.GameState),
		Period:        period,
		PeriodType:    periodType,
		Clock:         clock,
		TimeRemaining: clock, // Set both for frontend compatibility
		Venue: models.Venue{
			Name: scoreboard.Venue.Default,
		},
	}
	ret <- models.GameUpdate{
		OldState: game.CurrentState,
		NewState: newState,
	}
}

func (s NHLService) teamFromScheduleTeam(scheduleTeam nhl.NHLScheduleTeam) models.Team {
	//teamResp := s.Client.GetTeam(scheduleTeam.Abbrev).Teams[0]
	return models.Team{
		TeamName: scheduleTeam.PlaceName.Default,
		TeamCode: scheduleTeam.Abbrev,
		ExtID:    scheduleTeam.Abbrev,
		LeagueID: models.LeagueIdNHL,
		LogoURL:  scheduleTeam.Logo,
	}
}

func (s NHLService) gameFromSchedule(scheduleGame nhl.NHLScheduleResponseGame) models.Game {
	// Extract period information from periodDescriptor
	var period int
	var periodType string
	var clock string

	if scheduleGame.PeriodDescriptor.Number > 0 {
		period = scheduleGame.PeriodDescriptor.Number
		// Map period type: "REG" -> "REGULAR", "OT" -> "OVERTIME", etc.
		switch scheduleGame.PeriodDescriptor.PeriodType {
		case "REG":
			periodType = "REGULAR"
		case "OT":
			periodType = "OVERTIME"
		case "SO":
			periodType = "SHOOTOUT"
		default:
			periodType = "REGULAR"
		}
	} else if scheduleGame.GameState == "LIVE" {
		// Fallback to period 1 if periodDescriptor is missing but game is live
		period = 1
		periodType = "REGULAR"
		clock = "LIVE"
	}

	// Convert UTC time to local time for display
	localTime := scheduleGame.StartTimeUTC.Local()

	// Format game time for upcoming games
	var gameTimeDisplay string
	if scheduleGame.GameState == "PRE" || scheduleGame.GameState == "FUT" || scheduleGame.GameState == "OFF" {
		gameTimeDisplay = localTime.Format("Mon 3:04 PM")
	} else {
		// For live games, use clock if available, otherwise use "LIVE"
		if clock != "" {
			gameTimeDisplay = clock
		} else {
			gameTimeDisplay = "LIVE"
		}
	}

	return models.Game{
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team:  s.teamFromScheduleTeam(scheduleGame.HomeTeam),
				Score: scheduleGame.HomeTeam.Score,
				Statistics: models.TeamStats{
					Shots: scheduleGame.HomeTeam.Sog,
				},
			},
			Away: models.TeamState{
				Team:  s.teamFromScheduleTeam(scheduleGame.AwayTeam),
				Score: scheduleGame.AwayTeam.Score,
				Statistics: models.TeamStats{
					Shots: scheduleGame.AwayTeam.Sog,
				},
			},
			Status:        gameStatusFromScheduleGame(scheduleGame),
			FetchedAt:     time.Now(),
			Period:        period,
			PeriodType:    periodType,
			Clock:         gameTimeDisplay,
			TimeRemaining: gameTimeDisplay, // Set both for frontend compatibility
			Venue: models.Venue{
				Name: scheduleGame.Venue.Default,
			},
		},
		GameCode: strconv.Itoa(scheduleGame.ID),
		LeagueId: models.LeagueIdNHL,
		GameDetails: models.GameDetails{
			GameId:     strconv.Itoa(scheduleGame.ID),
			Season:     strconv.Itoa(scheduleGame.Season),
			SeasonType: strconv.Itoa(scheduleGame.GameType),
			GameDate:   localTime,
			GameTime:   localTime.Format("3:04 PM"),
			Timezone:   "Local",
		},
	}
}

func gameStatusFromScheduleGame(scheduleGame nhl.NHLScheduleResponseGame) models.GameStatus {
	return gameStatusFromGameState(scheduleGame.GameState)
}

func gameStatusFromGameState(gameState string) models.GameStatus {
	switch gameState {
	case STATUS_FINAL:
		return models.StatusEnded
	case STATUS_UPCOMING:
		return models.StatusUpcoming
	case STATUS_FUT:
		return models.StatusUpcoming
	case STATUS_OFF:
		return models.StatusUpcoming
	case STATUS_ACTIVE:
		return models.StatusActive
	default:
		return models.StatusActive
	}
}

func gameStatusFromStatusCode(statusCode string) models.GameStatus {
	if statusCode == "6" || statusCode == "7" {
		return models.StatusEnded
	}
	return models.StatusActive
}

func (s NHLService) GetEvents(update models.GameUpdate, ret chan []models.Event) {
	events := append(
		s.getGoalEvents(update.OldState.Home, update.NewState.Home, update.OldState.Away.Team),
		s.getGoalEvents(update.OldState.Away, update.NewState.Away, update.OldState.Home.Team)...,
	)
	ret <- events
}

func (s NHLService) getGoalEvents(oldState models.TeamState, newState models.TeamState, opponent models.Team) []models.Event {
	events := []models.Event{}
	diff := newState.Score - oldState.Score
	team := newState.Team

	for i := 0; i < diff; i++ {
		events = append(events, models.Event{
			TeamCode:     team.TeamCode,
			TeamName:     team.TeamName,
			TeamHash:     team.GetTeamHash(),
			LeagueId:     models.LeagueIdNHL,
			LeagueName:   s.GetLeagueName(),
			OpponentCode: opponent.TeamCode,
			OpponentName: opponent.TeamName,
			OpponentHash: opponent.GetTeamHash(),
		})
	}
	return events
}
