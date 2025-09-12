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
	// Use the actual CFL API to get schedule data
	logger.Info("Fetching CFL schedule data from official API")
	schedule := s.Client.GetCFLSchedule()

	// If the API returns empty data, log it and return empty schedule
	if len(schedule) == 0 {
		logger.Info("CFL API returned empty data - no games currently scheduled")
		return cfl.CFLScheduleResponse{}
	}

	return schedule
}

func (s CFLService) GetActiveGames(ret chan []models.Game) {
	schedule := s.getSchedule()
	var activeGames []models.Game

	logger.Infof("CFL GetActiveGames: Found %d rounds in schedule", len(schedule))

	for _, round := range schedule {
		logger.Infof("CFL GetActiveGames: Processing round %d (%s) with %d tournaments", round.ID, round.Name, len(round.Tournaments))
		for _, game := range round.Tournaments {
			gameStatus := gameStatusFromCFLGame(game)
			logger.Infof("CFL GetActiveGames: Game %d (%s vs %s) - Status: %s, Clock: %s, GameStatus: %d",
				game.ID, game.AwaySquad.ShortName, game.HomeSquad.ShortName, game.Status, game.Clock, gameStatus)

			if gameStatus == models.StatusActive {
				logger.Infof("CFL GetActiveGames: Adding active game %d to active games list", game.ID)
				activeGames = append(activeGames, s.gameFromCFLGame(game))
			}
		}
	}

	logger.Infof("CFL GetActiveGames: Returning %d active games", len(activeGames))
	ret <- activeGames
}

func (s CFLService) GetUpcomingGames(ret chan []models.Game) {
	schedule := s.getSchedule()
	var upcomingGames []models.Game

	logger.Infof("CFL GetUpcomingGames: Found %d rounds in schedule", len(schedule))

	for _, round := range schedule {
		logger.Infof("CFL GetUpcomingGames: Processing round %d (%s) with %d tournaments", round.ID, round.Name, len(round.Tournaments))
		for _, game := range round.Tournaments {
			gameStatus := gameStatusFromCFLGame(game)
			logger.Infof("CFL GetUpcomingGames: Game %d (%s vs %s) - Status: %s, Clock: %s, GameStatus: %d",
				game.ID, game.AwaySquad.ShortName, game.HomeSquad.ShortName, game.Status, game.Clock, gameStatus)

			if gameStatus == models.StatusUpcoming {
				logger.Infof("CFL GetUpcomingGames: Adding upcoming game %d to upcoming games list", game.ID)
				upcomingGames = append(upcomingGames, s.gameFromCFLGame(game))
			}
		}
	}

	logger.Infof("CFL GetUpcomingGames: Returning %d upcoming games", len(upcomingGames))
	ret <- upcomingGames
}

func (s CFLService) GetGameUpdate(game models.Game, ret chan models.GameUpdate) {
	// Use the actual Genius Sports API to get live game data
	logger.Info("Fetching live CFL game data from Genius Sports API")
	liveGame := s.Client.GetCFLLiveGame(game.GameCode)

	// Check if we got valid data from the API
	if liveGame.Data.BetGeniusFixtureID == "" {
		logger.Warnf("No live game data received for CFL game %s", game.GameCode)
		// Fall back to basic game state if no live data
		ret <- models.GameUpdate{
			OldState: game.CurrentState,
			NewState: game.CurrentState, // Keep existing state
		}
		return
	}

	// Extract detailed game information
	var period int
	var periodType string
	var clock string
	var details models.EventDetails

	logger.Infof("CFL GetGameUpdate: MatchStatus='%s', CurrentPhase='%s'", liveGame.Data.ScoreboardInfo.MatchStatus, liveGame.Data.ScoreboardInfo.CurrentPhase)

	if liveGame.Data.ScoreboardInfo.MatchStatus == STATUS_LIVE {
		// Extract period information from live stream data
		period = s.extractPeriodFromLiveStream(liveGame.Data.LiveStream.CurrentPlay.Phase)
		periodType = "QUARTER"
		clock = liveGame.Data.LiveStream.CurrentPlay.Clock

		// Extract detailed football-specific information from live stream
		details = models.EventDetails{
			// Extract down information from current play
			Down: liveGame.Data.LiveStream.CurrentPlay.DownNumber,
			// Extract distance information from current play
			Distance: liveGame.Data.LiveStream.CurrentPlay.YardsToGo,
			// Extract yard line information from current play
			YardLine: liveGame.Data.LiveStream.CurrentPlay.LineOfScrimmage,
			// Extract possession information from current play
			Possession: liveGame.Data.LiveStream.CurrentPlay.Possession,
		}

		logger.Infof("CFL GetGameUpdate: Live game details - Period=%d, Clock='%s', Down=%d, Distance=%d, YardLine=%d, Possession='%s', PlayType='%s', Formation='%s'",
			period, clock, details.Down, details.Distance, details.YardLine, details.Possession,
			liveGame.Data.LiveStream.CurrentPlay.PlayType, liveGame.Data.LiveStream.CurrentPlay.PlayFormation)
	} else {
		// For non-live games, use basic information
		period = 1
		periodType = "QUARTER"
		clock = "PRE-GAME"
		details = models.EventDetails{}
		logger.Infof("CFL GetGameUpdate: Non-live game - Period=%d, Clock='%s'", period, clock)
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
		Details:    details,
		Venue: models.Venue{
			Name: liveGame.Data.MatchInfo.VenueName,
		},
	}

	logger.Infof("CFL GetGameUpdate: New state - Status='%s', Period=%d, Clock='%s'", newState.Status, newState.Period, newState.Clock)

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
	// For now, return empty string to let the frontend handle fallback
	// The frontend will show team codes when logos fail to load
	// In the future, we can implement proper logo fetching from live game data
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

// Helper functions to extract football-specific data
func (s CFLService) extractDownFromLiveData(down interface{}) int {
	if down == nil {
		return 0
	}

	// Try to convert to int
	if downInt, ok := down.(int); ok {
		return downInt
	}

	// Try to convert from string
	if downStr, ok := down.(string); ok {
		if downInt, err := strconv.Atoi(downStr); err == nil {
			return downInt
		}
	}

	return 0
}

func (s CFLService) extractDistanceFromLiveData(yardsToGo interface{}) int {
	if yardsToGo == nil {
		return 0
	}

	// Try to convert to int
	if yardsInt, ok := yardsToGo.(int); ok {
		return yardsInt
	}

	// Try to convert from string
	if yardsStr, ok := yardsToGo.(string); ok {
		if yardsInt, err := strconv.Atoi(yardsStr); err == nil {
			return yardsInt
		}
	}

	return 0
}

func (s CFLService) extractYardLineFromLiveData(scoreboardInfo cfl.CFLScoreboardInfo) int {
	// For now, we'll return 0 as yard line information isn't directly available
	// In a real implementation, this might be calculated from other data
	// or available in a different part of the API response
	return 0
}

func (s CFLService) extractPossessionFromLiveData(possession string) string {
	// Return the possession team code
	// The possession field should contain the team code
	return possession
}

func (s CFLService) extractPeriodFromLiveStream(phase string) int {
	// Extract quarter number from phase string (e.g., "Q2" -> 2)
	if len(phase) >= 2 && phase[0] == 'Q' {
		if quarter, err := strconv.Atoi(phase[1:]); err == nil {
			return quarter
		}
	}
	return 1 // Default to quarter 1
}

func (s CFLService) gameFromCFLGame(cflGame cfl.CFLGame) models.Game {
	// Extract game details
	var period int
	var periodType string
	var clock string
	var venue models.Venue
	var details models.EventDetails

	// Set period information for CFL games
	if cflGame.Status == STATUS_LIVE {
		period = 1 // Default to quarter 1 for live games
		periodType = "QUARTER"
		clock = "LIVE"
	} else if cflGame.Status == STATUS_SCHEDULED {
		periodType = "QUARTER"
		clock = cflGame.Clock
	}

	// Set venue information if available
	// Note: CFLGame doesn't have venue info, but we can get it from live game data
	venue = models.Venue{
		Name: "", // Will be populated from live game data if available
	}

	// Set football-specific details
	if cflGame.Status == STATUS_LIVE {
		// For live games, set empty details that will be populated from live game data
		details = models.EventDetails{}
	} else {
		// For non-live games, empty details
		details = models.EventDetails{}
	}

	return models.Game{
		CurrentState: models.GameState{
			Home:       models.TeamState{Team: s.teamFromCFLTeam(cflGame.HomeSquad), Score: cflGame.HomeSquad.Score},
			Away:       models.TeamState{Team: s.teamFromCFLTeam(cflGame.AwaySquad), Score: cflGame.AwaySquad.Score},
			Status:     gameStatusFromCFLGame(cflGame),
			FetchedAt:  time.Now(),
			Period:     period,
			PeriodType: periodType,
			Clock:      clock,
			Venue:      venue,
			Details:    details,
		},
		GameCode: strconv.Itoa(cflGame.ID),
		LeagueId: models.LeagueIdCFL,
	}
}

func gameStatusFromCFLGame(cflGame cfl.CFLGame) models.GameStatus {
	logger.Infof("CFL Game Status Check: ID=%d, Status='%s', Clock='%s'", cflGame.ID, cflGame.Status, cflGame.Clock)

	switch cflGame.Status {
	case STATUS_COMPLETE:
		logger.Infof("CFL Game %d: Status COMPLETE -> StatusEnded", cflGame.ID)
		return models.StatusEnded
	case STATUS_SCHEDULED:
		logger.Infof("CFL Game %d: Status SCHEDULED -> StatusUpcoming", cflGame.ID)
		return models.StatusUpcoming
	case STATUS_LIVE:
		logger.Infof("CFL Game %d: Status LIVE -> StatusActive", cflGame.ID)
		return models.StatusActive
	default:
		// Check if game is currently live based on clock and other indicators
		if cflGame.Clock != "" && cflGame.Clock != "00:00" {
			logger.Infof("CFL Game %d: Default case with clock '%s' -> StatusActive", cflGame.ID, cflGame.Clock)
			return models.StatusActive
		}
		logger.Infof("CFL Game %d: Default case -> StatusUpcoming", cflGame.ID)
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

	// Emit a single scoring event per update if the score increased
	if diff > 0 {
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
