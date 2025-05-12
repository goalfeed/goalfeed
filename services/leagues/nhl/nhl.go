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
	newState := models.GameState{
		Home: models.TeamState{
			Team:  game.CurrentState.Home.Team,
			Score: scoreboard.HomeTeam.Score,
		},
		Away: models.TeamState{
			Team:  game.CurrentState.Away.Team,
			Score: scoreboard.AwayTeam.Score,
		},
		Status: gameStatusFromGameState(scoreboard.GameState),
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
	}
}

func (s NHLService) gameFromSchedule(scheduleGame nhl.NHLScheduleResponseGame) models.Game {
	return models.Game{
		CurrentState: models.GameState{
			Home:      models.TeamState{Team: s.teamFromScheduleTeam(scheduleGame.HomeTeam), Score: scheduleGame.HomeTeam.Score},
			Away:      models.TeamState{Team: s.teamFromScheduleTeam(scheduleGame.AwayTeam), Score: scheduleGame.AwayTeam.Score},
			Status:    gameStatusFromScheduleGame(scheduleGame),
			FetchedAt: time.Now(),
		},
		GameCode: strconv.Itoa(scheduleGame.ID),
		LeagueId: models.LeagueIdNHL,
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
		s.getGoalEvents(update.OldState.Home, update.NewState.Home),
		s.getGoalEvents(update.OldState.Away, update.NewState.Away)...,
	)
	ret <- events
}

func (s NHLService) getGoalEvents(oldState models.TeamState, newState models.TeamState) []models.Event {
	events := []models.Event{}
	diff := newState.Score - oldState.Score
	team := newState.Team
	opponent := oldState.Team // The opponent is the other team in the game state

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
