package cfl

import (
	"goalfeed/clients/leagues/cfl"
	"goalfeed/models"
	"goalfeed/utils"
	"strconv"
	"strings"
	"time"
)

const (
	STATUS_SCHEDULED = "scheduled"
	STATUS_COMPLETE  = "complete"
	STATUS_LIVE      = "live"
)

type CFLService struct {
	Client cfl.ICFLApiClient
}

var logger = utils.GetLogger()

func (s CFLService) GetLeagueName() string {
	return "CFL"
}

func (s CFLService) getSchedule() cfl.CFLScheduleResponse {
	return s.Client.GetCFLSchedule()
}

func (s CFLService) GetActiveGames(ret chan []models.Game) {
	schedule := s.getSchedule()
	var activeGames []models.Game

	for _, round := range schedule {
		for _, game := range round.Tournaments {
			if gameStatusFromCFLGame(game) == models.StatusActive {
				activeGames = append(activeGames, s.gameFromCFLGame(game))
			}
		}
	}
	ret <- activeGames
}

func (s CFLService) GetUpcomingGames(ret chan []models.Game) {
	schedule := s.getSchedule()
	var upcomingGames []models.Game

	for _, round := range schedule {
		for _, game := range round.Tournaments {
			if gameStatusFromCFLGame(game) == models.StatusUpcoming {
				upcomingGames = append(upcomingGames, s.gameFromCFLGame(game))
			}
		}
	}
	ret <- upcomingGames
}

func (s CFLService) GetGameUpdate(game models.Game, ret chan models.GameUpdate) {
	// For CFL, we need to get the live game data using the fixture ID
	// The game ID in our system should be the CFL fixture ID
	liveGame := s.Client.GetCFLLiveGame(game.GameCode)

	// Extract quarter information
	var period int
	var periodType string
	var clock string

	if liveGame.Data.ScoreboardInfo.MatchStatus == STATUS_LIVE {
		// CFL games have quarters, try to extract from the data
		period = 1 // Default to quarter 1 for live games
		periodType = "QUARTER"
		clock = "LIVE"
	}

	newState := models.GameState{
		Home: models.TeamState{
			Team:  game.CurrentState.Home.Team,
			Score: liveGame.Data.ScoreboardInfo.HomeScore,
		},
		Away: models.TeamState{
			Team:  game.CurrentState.Away.Team,
			Score: liveGame.Data.ScoreboardInfo.AwayScore,
		},
		Status:     gameStatusFromLiveGame(liveGame.Data.ScoreboardInfo.MatchStatus),
		Period:     period,
		PeriodType: periodType,
		Clock:      clock,
	}
	ret <- models.GameUpdate{
		OldState: game.CurrentState,
		NewState: newState,
	}
}

func (s CFLService) teamFromCFLTeam(cflTeam cfl.CFLTeam) models.Team {
	// Try to get logo from live game data if available
	logoURL := s.getTeamLogoFromLiveData(strconv.Itoa(cflTeam.ID))

	return models.Team{
		TeamName: cflTeam.Name,
		TeamCode: cflTeam.ShortName,
		ExtID:    strconv.Itoa(cflTeam.ID),
		LeagueID: models.LeagueIdCFL,
		LogoURL:  logoURL,
	}
}

func (s CFLService) getTeamLogoFromLiveData(teamID string) string {
	// Try to get logo from live game data
	// This is a simple approach - in a real implementation, you might want to cache this
	// or have a more sophisticated logo lookup system

	// For now, we'll use a fallback approach with known CFL team logos
	// The live game data has logos in CFLBrand.Logo and CFLTheme.Logo.SVG
	// but we need to make a request to get them, which might be expensive

	// Fallback to known CFL team logos based on team codes
	teamLogos := map[string]string{
		"93775":  "https://a.espncdn.com/i/teamlogos/cfl/500/bc.png",  // BC Lions
		"112939": "https://a.espncdn.com/i/teamlogos/cfl/500/cgy.png", // Calgary Stampeders
		"114347": "https://a.espncdn.com/i/teamlogos/cfl/500/edm.png", // Edmonton Elks
		"83579":  "https://a.espncdn.com/i/teamlogos/cfl/500/ham.png", // Hamilton Tiger-Cats
		"86680":  "https://a.espncdn.com/i/teamlogos/cfl/500/mtl.png", // Montreal Alouettes
		"88019":  "https://a.espncdn.com/i/teamlogos/cfl/500/ott.png", // Ottawa RedBlacks
		"106752": "https://a.espncdn.com/i/teamlogos/cfl/500/ssk.png", // Saskatchewan Roughriders
		"122345": "https://a.espncdn.com/i/teamlogos/cfl/500/tor.png", // Toronto Argonauts
		"110380": "https://a.espncdn.com/i/teamlogos/cfl/500/wpg.png", // Winnipeg Blue Bombers
	}

	if logo, exists := teamLogos[teamID]; exists {
		return logo
	}

	return ""
}

func (s CFLService) getTeamLogoFromLiveGameResponse(liveGame cfl.CFLLiveGameResponse, teamID string) string {
	// Extract logo from live game response
	// Check both home and away teams
	if liveGame.Data.MatchInfo.HomeTeam.CompetitorID == teamID {
		// Try to get logo from brand data
		if liveGame.Data.MatchInfo.HomeTeam.Details.Brand.Logo != "" {
			return liveGame.Data.MatchInfo.HomeTeam.Details.Brand.Logo
		}
		// Fallback to SVG logo
		if liveGame.Data.MatchInfo.HomeTeam.Details.Brand.Theme.Light.Logo.SVG != "" {
			return liveGame.Data.MatchInfo.HomeTeam.Details.Brand.Theme.Light.Logo.SVG
		}
	}

	if liveGame.Data.MatchInfo.AwayTeam.CompetitorID == teamID {
		// Try to get logo from brand data
		if liveGame.Data.MatchInfo.AwayTeam.Details.Brand.Logo != "" {
			return liveGame.Data.MatchInfo.AwayTeam.Details.Brand.Logo
		}
		// Fallback to SVG logo
		if liveGame.Data.MatchInfo.AwayTeam.Details.Brand.Theme.Light.Logo.SVG != "" {
			return liveGame.Data.MatchInfo.AwayTeam.Details.Brand.Theme.Light.Logo.SVG
		}
	}

	// Fallback to static logo mapping
	return s.getTeamLogoFromLiveData(teamID)
}

func (s CFLService) gameFromCFLGame(cflGame cfl.CFLGame) models.Game {
	return models.Game{
		CurrentState: models.GameState{
			Home:      models.TeamState{Team: s.teamFromCFLTeam(cflGame.HomeSquad), Score: cflGame.HomeSquad.Score},
			Away:      models.TeamState{Team: s.teamFromCFLTeam(cflGame.AwaySquad), Score: cflGame.AwaySquad.Score},
			Status:    gameStatusFromCFLGame(cflGame),
			FetchedAt: time.Now(),
		},
		GameCode: strconv.Itoa(cflGame.ID),
		LeagueId: models.LeagueIdCFL,
	}
}

func gameStatusFromCFLGame(cflGame cfl.CFLGame) models.GameStatus {
	switch cflGame.Status {
	case STATUS_COMPLETE:
		return models.StatusEnded
	case STATUS_SCHEDULED:
		return models.StatusUpcoming
	case STATUS_LIVE:
		return models.StatusActive
	default:
		// Check if game is currently live based on clock and other indicators
		if cflGame.Clock != "" && cflGame.Clock != "00:00" {
			return models.StatusActive
		}
		return models.StatusUpcoming
	}
}

func gameStatusFromLiveGame(matchStatus string) models.GameStatus {
	switch strings.ToLower(matchStatus) {
	case "prematch":
		return models.StatusUpcoming
	case "live", "inprogress":
		return models.StatusActive
	case "final", "complete":
		return models.StatusEnded
	default:
		return models.StatusActive
	}
}

func (s CFLService) GetEvents(update models.GameUpdate, ret chan []models.Event) {
	events := append(
		s.getTouchdownEvents(update.OldState.Home, update.NewState.Home, update.OldState.Away.Team),
		s.getTouchdownEvents(update.OldState.Away, update.NewState.Away, update.OldState.Home.Team)...,
	)
	ret <- events
}

func (s CFLService) getTouchdownEvents(oldState models.TeamState, newState models.TeamState, opponent models.Team) []models.Event {
	events := []models.Event{}
	diff := newState.Score - oldState.Score
	team := newState.Team

	// In CFL, we'll treat score changes as touchdowns (6 points each)
	// We'll create events for each 6-point increment
	for i := 0; i < diff; i += 6 {
		events = append(events, models.Event{
			TeamCode:     team.TeamCode,
			TeamName:     team.TeamName,
			TeamHash:     team.GetTeamHash(),
			LeagueId:     models.LeagueIdCFL,
			LeagueName:   s.GetLeagueName(),
			OpponentCode: opponent.TeamCode,
			OpponentName: opponent.TeamName,
			OpponentHash: opponent.GetTeamHash(),
		})
	}
	return events
}
