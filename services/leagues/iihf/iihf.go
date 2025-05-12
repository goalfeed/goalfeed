package iihf

import (
	"goalfeed/clients/leagues/iihf"
	"goalfeed/models"
	"time"
)

type IIHFService struct {
	Client iihf.IIIHFApiClient
}

const STATUS_UPCOMING = "UPCOMING"
const STATUS_ACTIVE = "LIVE"
const STATUS_FINAL = "FINAL"

// const IIHF_LEAGUE_ID = 4

func (s IIHFService) getSchedule() iihf.IIHFScheduleResponse {

	//todo implement caching
	//todo support multiple active events
	//todo support some method of determining active events programmatically
	return s.Client.GetIIHFSchedule("503")
}
func (s IIHFService) GetLeagueName() string {
	return "IIHF"
}

// GetActiveGames Returns active IIHFGames
func (s IIHFService) GetActiveGames(ret chan []models.Game) {
	schedule := s.getSchedule()
	var activeGames []models.Game
	for _, game := range schedule {
		gameFromSchedule(game)
		if gameStatusFromScheduleGame(game) == models.StatusActive {
			activeGames = append(activeGames, gameFromSchedule(game))
		}
	}
	ret <- activeGames
}

// GetActiveGames Returns a GameUpdate
func (s IIHFService) GetGameUpdate(game models.Game, ret chan models.GameUpdate) {
	scoreboard := s.Client.GetIIHFScoreBoard(game.GameCode)
	newState := models.GameState{
		Home: models.TeamState{
			Team:  game.CurrentState.Home.Team,
			Score: scoreboard.CurrentScore.Home,
		},
		Away: models.TeamState{
			Team:  game.CurrentState.Away.Team,
			Score: scoreboard.CurrentScore.Away,
		},
		Status: game.CurrentState.Status, //TODO: Update using scoreboard status
	}
	ret <- models.GameUpdate{
		OldState: game.CurrentState,
		NewState: newState,
	}
}

func teamFromScheduleTeam(scheduleTeam iihf.IIHFScheduleTeam) models.Team {

	// todo store/retrieve from DB
	// todo fill out model
	team := models.Team{
		TeamName: scheduleTeam.TeamCode,
		TeamCode: scheduleTeam.TeamCode,
		ExtID:    scheduleTeam.TeamCode,
		LeagueID: models.LeagueIdIIHF,
	}
	return team

}
func gameFromSchedule(scheduleGame iihf.IIHFScheduleResponseGame) models.Game {

	return models.Game{
		CurrentState: models.GameState{
			Home:      models.TeamState{Team: teamFromScheduleTeam(scheduleGame.HomeTeam), Score: scheduleGame.HomeTeam.Points},
			Away:      models.TeamState{Team: teamFromScheduleTeam(scheduleGame.GuestTeam), Score: scheduleGame.GuestTeam.Points},
			Status:    gameStatusFromScheduleGame(scheduleGame),
			FetchedAt: time.Now(),
		},
		GameCode: scheduleGame.GameID,
		LeagueId: models.LeagueIdIIHF,
	}
}
func gameStatusFromScheduleGame(scheduleGame iihf.IIHFScheduleResponseGame) models.GameStatus {
	switch scheduleGame.Status {
	case STATUS_ACTIVE:
		return models.StatusActive
	case STATUS_FINAL:
		return models.StatusEnded
	case STATUS_UPCOMING:
		return models.StatusUpcoming
	default:
		//todo log error
		return models.StatusActive
	}
}
func (s IIHFService) GetEvents(update models.GameUpdate, ret chan []models.Event) {
	events := append(
		s.getGoalEvents(update.OldState.Home, update.NewState.Home, update.OldState.Away.Team),
		s.getGoalEvents(update.OldState.Away, update.NewState.Away, update.OldState.Home.Team)...,
	)
	ret <- events
}
func (s IIHFService) getGoalEvents(oldState models.TeamState, newState models.TeamState, opponent models.Team) []models.Event {
	events := []models.Event{}
	diff := newState.Score - oldState.Score
	if diff <= 0 {
		return events
	}
	team := newState.Team

	for i := 0; i < diff; i++ {
		events = append(events, models.Event{
			TeamCode:     team.TeamCode,
			TeamName:     team.TeamName,
			TeamHash:     team.GetTeamHash(),
			LeagueId:     models.LeagueIdIIHF,
			LeagueName:   s.GetLeagueName(),
			OpponentCode: opponent.TeamCode,
			OpponentName: opponent.TeamName,
			OpponentHash: opponent.GetTeamHash(),
		})
	}
	return events
}
