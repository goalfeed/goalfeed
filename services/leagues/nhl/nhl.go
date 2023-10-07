package nhl

import (
	"encoding/json"
	"fmt"
	"goalfeed/clients/leagues/nhl"
	"goalfeed/models"
	"goalfeed/targets/database"
	"goalfeed/utils"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type NHLService struct {
	Client nhl.INHLApiClient
}

const STATUS_UPCOMING = "Preview"
const STATUS_ACTIVE = "Live"
const STATUS_FINAL = "Final"

// const NHL_LEAGUE_ID = 4
var logger = utils.GetLogger()

func (s NHLService) getSchedule() nhl.NHLScheduleResponse {

	//todo implement caching
	//todo support multiple active events
	//todo support some method of determining active events programmatically
	return s.Client.GetNHLSchedule()
}
func (s NHLService) GetLeagueName() string {
	return "NHL"
}

// GetActiveGames Returns active NHLGames
func (s NHLService) GetActiveGames(ret chan []models.Game) {
	schedule := s.getSchedule()
	var activeGames []models.Game

	for _, date := range schedule.Dates {
		for _, game := range date.Games {
			tmpGame := s.gameFromSchedule(game)
			_ = tmpGame
			if gameStatusFromScheduleGame(game) == models.StatusActive {
				activeGames = append(activeGames, s.gameFromSchedule(game))
			}
		}
	}
	ret <- activeGames
}

// GetActiveGames Returns a GameUpdate
func (s NHLService) GetGameUpdate(game models.Game, ret chan models.GameUpdate) {
	fullCheckLotto := rand.Intn(180)
	if game.CurrentState.ExtTimestamp != "" && fullCheckLotto != 1 {
		s.getGameUpdateFromDiffPatch(game, ret)
		// s.getGameUpdateFromScoreboard(game, ret)
	} else {
		s.getGameUpdateFromScoreboard(game, ret)
	}
}
func fudgeTimestamp(extTimestamp string) string {

	pieces := strings.Split(extTimestamp, "_")
	oldTimeInt, _ := strconv.Atoi(pieces[1])
	newTimeInt := oldTimeInt - 10
	_ = pieces
	newTime := fmt.Sprintf("%s_%06d", pieces[0], newTimeInt)
	return newTime

}

func (s NHLService) getGameUpdateFromDiffPatch(game models.Game, ret chan models.GameUpdate) {

	diff, err := s.Client.GetDiffPatch(game.GameCode, fudgeTimestamp(game.CurrentState.ExtTimestamp))
	if err != nil {
		s.getGameUpdateFromScoreboard(game, ret)
		return
	}
	timestampPath := "/metaData/timeStamp"
	homeGoalPath := "/liveData/linescore/teams/home/goals"
	awayGoalPath := "/liveData/linescore/teams/away/goals"
	statusCodePath := "/gameData/status/statusCode"
	var extTimestamp string
	var homeScore int
	var awayScore int
	var statusCode string
	var status models.GameStatus

	for _, set := range diff {
		for _, item := range set.Diff {
			logger.Debug(fmt.Sprintf("Path: %s", item.Path))
			if item.Path == timestampPath {
				json.Unmarshal(item.Value, &extTimestamp)
			} else if item.Path == homeGoalPath {
				logger.Info(fmt.Sprintf("Home score change - %s", game.CurrentState.Home.Team.TeamName))
				json.Unmarshal(item.Value, &homeScore)
			} else if item.Path == awayGoalPath {
				logger.Info(fmt.Sprintf("Away score change - %s", game.CurrentState.Away.Team.TeamName))
				json.Unmarshal(item.Value, &awayScore)
			} else if item.Path == statusCodePath {
				logger.Info("Status Code")
				json.Unmarshal(item.Value, &statusCode)
			}
		}
	}

	if homeScore == 0 {
		homeScore = game.CurrentState.Home.Score
	}
	if awayScore == 0 {
		awayScore = game.CurrentState.Away.Score
	}
	if extTimestamp == "" {
		extTimestamp = game.CurrentState.ExtTimestamp
	}
	if statusCode == "" {
		status = game.CurrentState.Status
	} else {
		status = gameStatusFromStatusCode(statusCode)
	}

	newState := models.GameState{
		Home: models.TeamState{
			Team:  game.CurrentState.Home.Team,
			Score: homeScore,
		},
		Away: models.TeamState{
			Team:  game.CurrentState.Away.Team,
			Score: awayScore,
		},
		Status:       status,
		ExtTimestamp: extTimestamp,
	}

	ret <- models.GameUpdate{
		OldState: game.CurrentState,
		NewState: newState,
	}
}

func (s NHLService) getGameUpdateFromScoreboard(game models.Game, ret chan models.GameUpdate) {
	scoreboard := s.Client.GetNHLScoreBoard(game.GameCode)
	newState := models.GameState{
		Home: models.TeamState{
			Team:  game.CurrentState.Home.Team,
			Score: scoreboard.LiveData.Linescore.Teams.Home.Goals,
		},
		Away: models.TeamState{
			Team:  game.CurrentState.Away.Team,
			Score: scoreboard.LiveData.Linescore.Teams.Away.Goals,
		},
		Status:       gameStatusFromStatusCode(scoreboard.GameData.Status.StatusCode),
		ExtTimestamp: scoreboard.MetaData.TimeStamp,
	}
	ret <- models.GameUpdate{
		OldState: game.CurrentState,
		NewState: newState,
	}
}

func (s NHLService) teamFromScheduleTeam(scheduleTeam nhl.NHLScheduleTeam) models.Team {

	// todo store/retrieve from DB
	// todo fill out model
	teamResp := s.Client.GetTeam(scheduleTeam.Team.Link).Teams[0]
	team := models.Team{
		TeamName: teamResp.Name,
		TeamCode: teamResp.Abbreviation,
		ExtID:    teamResp.Abbreviation,
		LeagueID: models.LeagueIdNHL,
	}
	return database.GetOrCreateTeam(team)

}
func (s NHLService) gameFromSchedule(scheduleGame nhl.NHLScheduleResponseGame) models.Game {

	return models.Game{
		CurrentState: models.GameState{
			Home:      models.TeamState{Team: s.teamFromScheduleTeam(scheduleGame.Teams.Home), Score: scheduleGame.Teams.Home.Score},
			Away:      models.TeamState{Team: s.teamFromScheduleTeam(scheduleGame.Teams.Away), Score: scheduleGame.Teams.Away.Score},
			Status:    gameStatusFromScheduleGame(scheduleGame),
			FetchedAt: time.Now(),
		},
		GameCode: strconv.Itoa(scheduleGame.GamePk),
		LeagueId: models.LeagueIdNHL,
	}
}
func gameStatusFromScheduleGame(scheduleGame nhl.NHLScheduleResponseGame) models.GameStatus {
	switch scheduleGame.Status.AbstractGameState {
	case STATUS_FINAL:
		return models.StatusEnded
	case STATUS_UPCOMING:
		return models.StatusUpcoming
	case STATUS_ACTIVE:
		return models.StatusActive
	default:
		return models.StatusActive
	}
}
func gameStatusFromStatusCode(statusCode string) models.GameStatus {
	switch statusCode {
	case "6":
	case "7":
		return models.StatusEnded
	default:
		return models.StatusActive
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
	if diff <= 0 {
		return events
	}
	team := newState.Team
	for i := 0; i < diff; i++ {
		events = append(events, models.Event{
			TeamCode:   team.TeamCode,
			TeamName:   team.TeamName,
			TeamHash:   team.GetTeamHash(),
			LeagueId:   models.LeagueIdNHL,
			LeagueName: s.GetLeagueName(),
		})
	}
	return events

}
