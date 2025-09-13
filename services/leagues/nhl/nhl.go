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

func (s NHLService) GetActiveGames(ret chan []models.Game) {
	schedule := s.getSchedule()
	var activeGames []models.Game

	for _, date := range schedule.GameWeek {
		for _, game := range date.Games {
			if gameStatusFromScheduleGame(game) == models.StatusActive {
				activeGames = append(activeGames, s.gameFromSchedule(game))
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

func (s NHLService) getGameUpdateFromScoreboard(game models.Game, ret chan models.GameUpdate) {
	scoreboard := s.Client.GetNHLScoreBoard(game.GameCode)

	// Extract period information from game state
	var period int
	var periodType string
	var clock string

	// For NHL, we can extract basic info from the game state
	if scoreboard.GameState == "LIVE" {
		period = 1 // Default to period 1 for live games
		periodType = "REGULAR"
		clock = "LIVE"
	}

	newState := models.GameState{
		Home: models.TeamState{
			Team:  game.CurrentState.Home.Team,
			Score: scoreboard.HomeTeam.Score,
		},
		Away: models.TeamState{
			Team:  game.CurrentState.Away.Team,
			Score: scoreboard.AwayTeam.Score,
		},
		Status:     gameStatusFromGameState(scoreboard.GameState),
		Period:     period,
		PeriodType: periodType,
		Clock:      clock,
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
	// Extract period information
	var period int
	var periodType string
	var clock string

	if scheduleGame.GameState == "LIVE" {
		period = 1 // Default to period 1 for live games
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
		gameTimeDisplay = clock
	}

	return models.Game{
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team:  s.teamFromScheduleTeam(scheduleGame.HomeTeam),
				Score: scheduleGame.HomeTeam.Score,
			},
			Away: models.TeamState{
				Team:  s.teamFromScheduleTeam(scheduleGame.AwayTeam),
				Score: scheduleGame.AwayTeam.Score,
			},
			Status:     gameStatusFromScheduleGame(scheduleGame),
			FetchedAt:  time.Now(),
			Period:     period,
			PeriodType: periodType,
			Clock:      gameTimeDisplay,
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
