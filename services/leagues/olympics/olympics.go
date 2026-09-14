package olympics

import (
	"fmt"
	"goalfeed/clients/leagues/olympics"
	"goalfeed/models"
	"strings"
	"time"
)

const (
	statusUpcoming = "UPCOMING"
	statusLive     = "LIVE"
	statusActive   = "ACTIVE"
	statusFinal    = "FINAL"
	statusEnded    = "ENDED"
)

// OlympicHockeyService implements ILeagueService for Olympic men's or women's ice hockey
type OlympicHockeyService struct {
	Client     olympics.IOlympicsApiClient
	GenderCode string // "M" or "W"
	LeagueId   models.League
	LeagueName string
}

// NewMensHockeyService returns a service for Olympic men's hockey
func NewMensHockeyService(client olympics.IOlympicsApiClient) OlympicHockeyService {
	return OlympicHockeyService{
		Client:     client,
		GenderCode: olympics.GenderMen,
		LeagueId:   models.LeagueIdOlympicMensHockey,
		LeagueName: "Olympic Men's Hockey",
	}
}

// NewWomensHockeyService returns a service for Olympic women's hockey
func NewWomensHockeyService(client olympics.IOlympicsApiClient) OlympicHockeyService {
	return OlympicHockeyService{
		Client:     client,
		GenderCode: olympics.GenderWomen,
		LeagueId:   models.LeagueIdOlympicWomensHockey,
		LeagueName: "Olympic Women's Hockey",
	}
}

func (s OlympicHockeyService) GetLeagueName() string {
	return s.LeagueName
}

func (s OlympicHockeyService) getStartList() (olympics.StartListResponse, error) {
	return s.Client.GetStartList(s.GenderCode)
}

func gameCode(phaseCode, eventUnitCode string) string {
	return phaseCode + "_" + eventUnitCode
}

func parseGameCode(gameCode string) (phaseCode, eventUnitCode string) {
	idx := strings.Index(gameCode, "_")
	if idx < 0 {
		return "", gameCode
	}
	return gameCode[:idx], gameCode[idx+1:]
}

func (s OlympicHockeyService) teamFromCode(code, name string) models.Team {
	if name == "" {
		name = code
	}
	return models.Team{
		TeamName: name,
		TeamCode: code,
		ExtID:    code,
		LeagueID: int(s.LeagueId),
		LogoURL:  "",
	}
}

func (s OlympicHockeyService) gameFromEntry(entry olympics.StartListEntry, homeScore, awayScore int, status models.GameStatus) models.Game {
	homeCode := entry.HomeTeamCode
	awayCode := entry.AwayTeamCode
	homeName := entry.HomeTeam
	awayName := entry.AwayTeam
	if homeCode == "" || awayCode == "" {
		if len(entry.Participants) >= 2 {
			homeCode = entry.Participants[0].Code
			awayCode = entry.Participants[1].Code
			homeName = entry.Participants[0].Name
			awayName = entry.Participants[1].Name
		} else {
			homeCode = "HOME"
			awayCode = "AWAY"
		}
	}
	gc := gameCode(entry.PhaseCode, entry.EventUnitCode)
	return models.Game{
		GameCode: gc,
		LeagueId: s.LeagueId,
		CurrentState: models.GameState{
			Home:      models.TeamState{Team: s.teamFromCode(homeCode, homeName), Score: homeScore},
			Away:      models.TeamState{Team: s.teamFromCode(awayCode, awayName), Score: awayScore},
			Status:    status,
			FetchedAt: time.Now(),
		},
	}
}

func statusFromString(st string) models.GameStatus {
	switch strings.ToUpper(st) {
	case statusLive, statusActive:
		return models.StatusActive
	case statusFinal, statusEnded:
		return models.StatusEnded
	case statusUpcoming:
		return models.StatusUpcoming
	default:
		return models.StatusUpcoming
	}
}

func scoreFromEntry(entry olympics.StartListEntry) (home, away int) {
	if entry.Result != nil {
		home = entry.Result.HomeScore
		if home == 0 {
			home = entry.Result.ScoreHome
		}
		away = entry.Result.AwayScore
		if away == 0 {
			away = entry.Result.ScoreAway
		}
	}
	return home, away
}

// GetActiveGames returns games that are currently live/active
func (s OlympicHockeyService) GetActiveGames(ret chan []models.Game) {
	list, err := s.getStartList()
	if err != nil {
		ret <- nil
		return
	}
	scores, _ := s.Client.GetScores(s.GenderCode)
	scoreMap := scoresToMap(scores)
	var active []models.Game
	for _, entry := range list {
		st := statusFromString(entry.Status)
		if st != models.StatusActive {
			continue
		}
		home, away := scoreFromEntry(entry)
		if home == 0 && away == 0 {
			if scores, ok := scoreMap[gameCode(entry.PhaseCode, entry.EventUnitCode)]; ok {
				home, away = scores[0], scores[1]
			}
		}
		active = append(active, s.gameFromEntry(entry, home, away, st))
	}
	ret <- active
}

// GetUpcomingGames returns games that have not started
func (s OlympicHockeyService) GetUpcomingGames(ret chan []models.Game) {
	list, err := s.getStartList()
	if err != nil {
		ret <- nil
		return
	}
	var upcoming []models.Game
	for _, entry := range list {
		if statusFromString(entry.Status) != models.StatusUpcoming {
			continue
		}
		home, away := scoreFromEntry(entry)
		upcoming = append(upcoming, s.gameFromEntry(entry, home, away, models.StatusUpcoming))
	}
	ret <- upcoming
}

// GetGamesByDate is not supported by the Olympics API (no date filter in startList)
func (s OlympicHockeyService) GetGamesByDate(date string, ret chan []models.Game) {
	ret <- []models.Game{}
}

func scoresToMap(scores olympics.ScoresResponse) map[string][2]int {
	out := make(map[string][2]int)
	for _, u := range scores.EventUnits {
		home := u.HomeScore
		if home == 0 {
			home = u.ScoreHome
		}
		away := u.AwayScore
		if away == 0 {
			away = u.ScoreAway
		}
		out[gameCode(u.PhaseCode, u.EventUnitCode)] = [2]int{home, away}
	}
	// Single game response
	if scores.PhaseCode != "" && scores.UnitCode != "" {
		home := scores.HomeScore
		if home == 0 {
			home = scores.ScoreHome
		}
		away := scores.AwayScore
		if away == 0 {
			away = scores.ScoreAway
		}
		out[gameCode(scores.PhaseCode, scores.UnitCode)] = [2]int{home, away}
	}
	return out
}

// GetGameUpdate fetches current score for the game and returns old vs new state
func (s OlympicHockeyService) GetGameUpdate(game models.Game, ret chan models.GameUpdate) {
	phaseCode, eventUnitCode := parseGameCode(game.GameCode)
	scores, err := s.Client.GetScoresForGame(s.GenderCode, phaseCode, eventUnitCode)
	if err != nil {
		ret <- models.GameUpdate{OldState: game.CurrentState, NewState: game.CurrentState}
		return
	}
	home, away := 0, 0
	if scores.PhaseCode != "" && scores.UnitCode != "" {
		home = scores.HomeScore
		if home == 0 {
			home = scores.ScoreHome
		}
		away = scores.AwayScore
		if away == 0 {
			away = scores.ScoreAway
		}
	}
	newState := models.GameState{
		Home:      models.TeamState{Team: game.CurrentState.Home.Team, Score: home},
		Away:      models.TeamState{Team: game.CurrentState.Away.Team, Score: away},
		Status:    game.CurrentState.Status,
		FetchedAt: time.Now(),
	}
	ret <- models.GameUpdate{
		OldState: game.CurrentState,
		NewState: newState,
	}
}

// GetEvents produces goal events from score diff (old vs new)
func (s OlympicHockeyService) GetEvents(update models.GameUpdate, ret chan []models.Event) {
	events := append(
		s.goalEvents(update.OldState.Home, update.NewState.Home, update.OldState.Away.Team),
		s.goalEvents(update.OldState.Away, update.NewState.Away, update.OldState.Home.Team)...,
	)
	ret <- events
}

func (s OlympicHockeyService) goalEvents(oldState, newState models.TeamState, opponent models.Team) []models.Event {
	var out []models.Event
	diff := newState.Score - oldState.Score
	if diff <= 0 {
		return out
	}
	team := newState.Team
	for i := 0; i < diff; i++ {
		out = append(out, models.Event{
			TeamCode:     team.TeamCode,
			TeamName:     team.TeamName,
			TeamHash:     team.GetTeamHash(),
			LeagueId:     int(s.LeagueId),
			LeagueName:   s.GetLeagueName(),
			Type:         models.EventTypeGoal,
			Description:  fmt.Sprintf("Goal! %s", team.TeamName),
			OpponentCode: opponent.TeamCode,
			OpponentName: opponent.TeamName,
			OpponentHash: opponent.GetTeamHash(),
			Timestamp:    time.Now(),
		})
	}
	return out
}
