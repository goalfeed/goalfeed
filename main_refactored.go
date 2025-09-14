package main

import (
	"fmt"

	"goalfeed/models"
	"goalfeed/services/leagues"
	"goalfeed/targets/homeassistant"
	"goalfeed/targets/memoryStore"
	webApi "goalfeed/web/api"
)

// GameChecker handles the complex logic of checking and updating games
type GameChecker struct {
	gameKey string
	game    models.Game
	service leagues.ILeagueService
}

// NewGameChecker creates a new game checker instance
func NewGameChecker(gameKey string) (*GameChecker, error) {
	game, err := memoryStore.GetGameByGameKey(gameKey)
	if err != nil {
		logger.Error(err.Error())
		logger.Error(fmt.Sprintf("[%s] Game not found, skipping", gameKey))
		memoryStore.DeleteActiveGameKey(gameKey)
		needRefresh = true
		return nil, err
	}

	service := leagueServices[int(game.LeagueId)]
	return &GameChecker{
		gameKey: gameKey,
		game:    game,
		service: service,
	}, nil
}

// LogGameInfo logs the current game state
func (gc *GameChecker) LogGameInfo() {
	logger.Info(fmt.Sprintf("[%s - %s %d @ %s %d] Checking",
		gc.service.GetLeagueName(),
		gc.game.CurrentState.Away.Team.TeamCode,
		gc.game.CurrentState.Away.Score,
		gc.game.CurrentState.Home.Team.TeamCode,
		gc.game.CurrentState.Home.Score))
}

// FetchGameUpdate retrieves the latest game update from the service
func (gc *GameChecker) FetchGameUpdate() (models.GameUpdate, error) {
	gc.game.IsFetching = true

	updateChan := make(chan models.GameUpdate)
	go gc.service.GetGameUpdate(gc.game, updateChan)
	update := <-updateChan

	return update, nil
}

// FetchEvents retrieves events for the game update
func (gc *GameChecker) FetchEvents(update models.GameUpdate) {
	eventChan := make(chan []models.Event)
	go gc.service.GetEvents(update, eventChan)
	go fireGoalEvents(eventChan, gc.game)
}

// HasMeaningfulChange checks if the game update contains meaningful changes
func (gc *GameChecker) HasMeaningfulChange(update models.GameUpdate) bool {
	ns := update.NewState
	os := gc.game.CurrentState

	if ns.Status != os.Status {
		return true
	}
	if ns.Home.Score != os.Home.Score || ns.Away.Score != os.Away.Score {
		return true
	}
	if ns.Period != os.Period {
		return true
	}
	if ns.Clock != os.Clock {
		return true
	}
	if ns.ExtTimestamp != "" && ns.ExtTimestamp != os.ExtTimestamp {
		return true
	}

	return false
}

// HandleNoChange handles the case when no meaningful changes are detected
func (gc *GameChecker) HandleNoChange() {
	gc.game.IsFetching = false
	memoryStore.SetGame(gc.game)
}

// HandlePeriodChange handles period change notifications
func (gc *GameChecker) HandlePeriodChange(oldPeriod int, update models.GameUpdate) {
	if gc.game.CurrentState.Period > oldPeriod {
		// Period started
		go homeassistant.SendPeriodUpdate(gc.game, models.EventTypePeriodStart)
		logger.Info(fmt.Sprintf("[%s - %s @ %s] Period %d started",
			gc.service.GetLeagueName(),
			gc.game.CurrentState.Away.Team.TeamCode,
			gc.game.CurrentState.Home.Team.TeamCode,
			gc.game.CurrentState.Period))
	}
}

// HandleStatusChange handles game status change notifications
func (gc *GameChecker) HandleStatusChange(oldStatus models.GameStatus, oldPeriod int, update models.GameUpdate) {
	switch gc.game.CurrentState.Status {
	case models.StatusActive:
		// Announce start only on a real edge with advancement
		advanced := (update.NewState.ExtTimestamp != update.OldState.ExtTimestamp) ||
			(gc.game.CurrentState.Period > oldPeriod) ||
			(gc.game.CurrentState.Clock != "")
		if oldStatus == models.StatusUpcoming && advanced {
			go homeassistant.SendPeriodUpdate(gc.game, models.EventTypeGameStart)
			logger.Info(fmt.Sprintf("[%s - %s @ %s] Game started",
				gc.service.GetLeagueName(),
				gc.game.CurrentState.Away.Team.TeamCode,
				gc.game.CurrentState.Home.Team.TeamCode))
		}
	case models.StatusEnded:
		go homeassistant.SendPeriodUpdate(gc.game, models.EventTypeGameEnd)
		// Reset sensors at end of game
		go homeassistant.PublishEndOfGameReset(gc.game)
		logger.Info(fmt.Sprintf("[%s - %s @ %s] Game ended",
			gc.service.GetLeagueName(),
			gc.game.CurrentState.Away.Team.TeamCode,
			gc.game.CurrentState.Home.Team.TeamCode))
	}
}

// HandleGameEnd handles cleanup when a game ends
func (gc *GameChecker) HandleGameEnd() {
	memoryStore.DeleteActiveGame(gc.game)
	memoryStore.DeleteActiveGameKey(gc.game.GetGameKey())
}

// HandleGameUpdate handles normal game state updates
func (gc *GameChecker) HandleGameUpdate() {
	gc.game.IsFetching = false
	memoryStore.SetGame(gc.game)

	// Send enhanced game update to Home Assistant
	go homeassistant.SendGameUpdate(gc.game)
	// Publish team-first sensors
	go homeassistant.PublishTeamSensors(gc.game)
	// Broadcast game update to web clients
	webApi.BroadcastGameUpdate(gc.game)
}

// CheckGameRefactored is the refactored version of checkGame
func CheckGameRefactored(gameKey string) {
	// Create game checker
	checker, err := NewGameChecker(gameKey)
	if err != nil {
		return
	}

	// Log current game info
	checker.LogGameInfo()

	// Fetch game update
	update, err := checker.FetchGameUpdate()
	if err != nil {
		return
	}

	// Fetch events
	checker.FetchEvents(update)

	// Check for meaningful changes
	if !checker.HasMeaningfulChange(update) {
		checker.HandleNoChange()
		return
	}

	// Store old state for comparison
	oldPeriod := checker.game.CurrentState.Period
	oldStatus := checker.game.CurrentState.Status

	// Update game state
	checker.game.CurrentState = update.NewState

	// Handle period changes
	checker.HandlePeriodChange(oldPeriod, update)

	// Handle status changes
	checker.HandleStatusChange(oldStatus, oldPeriod, update)

	// Handle game end or normal update
	if checker.game.CurrentState.Status == models.StatusEnded {
		checker.HandleGameEnd()
	} else {
		checker.HandleGameUpdate()
	}
}
