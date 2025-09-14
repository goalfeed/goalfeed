package main

import (
	"testing"

	"goalfeed/models"
	"goalfeed/targets/memoryStore"

	"github.com/stretchr/testify/assert"
)

func TestNewGameChecker(t *testing.T) {
	setupTest(t)

	// Test with non-existent game
	checker, err := NewGameChecker("non-existent-game")
	assert.Error(t, err)
	assert.Nil(t, checker)

	// Test with existing game
	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")
	memoryStore.AppendActiveGame(game)

	checker, err = NewGameChecker(game.GetGameKey())
	assert.NoError(t, err)
	assert.NotNil(t, checker)
	assert.Equal(t, game.GetGameKey(), checker.gameKey)
	assert.Equal(t, game.GameCode, checker.game.GameCode)
}

func TestGameChecker_LogGameInfo(t *testing.T) {
	setupTest(t)

	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")
	game.CurrentState.Home.Score = 2
	game.CurrentState.Away.Score = 1
	memoryStore.AppendActiveGame(game)

	checker, err := NewGameChecker(game.GetGameKey())
	assert.NoError(t, err)

	// Test that LogGameInfo doesn't panic
	assert.NotPanics(t, func() {
		checker.LogGameInfo()
	})
}

func TestGameChecker_FetchGameUpdate(t *testing.T) {
	setupTest(t)

	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")
	memoryStore.AppendActiveGame(game)

	checker, err := NewGameChecker(game.GetGameKey())
	assert.NoError(t, err)

	// Test that FetchGameUpdate doesn't panic
	assert.NotPanics(t, func() {
		update, err := checker.FetchGameUpdate()
		assert.NoError(t, err)
		assert.NotNil(t, update)
	})
}

func TestGameChecker_FetchEvents(t *testing.T) {
	setupTest(t)

	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")
	memoryStore.AppendActiveGame(game)

	checker, err := NewGameChecker(game.GetGameKey())
	assert.NoError(t, err)

	update := models.GameUpdate{
		OldState: game.CurrentState,
		NewState: game.CurrentState,
	}

	// Test that FetchEvents doesn't panic
	assert.NotPanics(t, func() {
		checker.FetchEvents(update)
	})
}

func TestGameChecker_HasMeaningfulChange(t *testing.T) {
	setupTest(t)

	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")
	memoryStore.AppendActiveGame(game)

	checker, err := NewGameChecker(game.GetGameKey())
	assert.NoError(t, err)

	// Test with no changes
	update := models.GameUpdate{
		OldState: game.CurrentState,
		NewState: game.CurrentState,
	}
	assert.False(t, checker.HasMeaningfulChange(update))

	// Test with status change
	oldState := checker.game.CurrentState
	newState := models.GameState{
		Status:       models.StatusEnded, // Change to a different status
		Home:         oldState.Home,
		Away:         oldState.Away,
		Period:       oldState.Period,
		Clock:        oldState.Clock,
		ExtTimestamp: oldState.ExtTimestamp,
	}
	update = models.GameUpdate{
		OldState: oldState,
		NewState: newState,
	}

	assert.True(t, checker.HasMeaningfulChange(update))

	// Test with score change
	oldState = checker.game.CurrentState
	newState = models.GameState{
		Status: oldState.Status,
		Home: models.TeamState{
			Team:  oldState.Home.Team,
			Score: 1,
		},
		Away:         oldState.Away,
		Period:       oldState.Period,
		Clock:        oldState.Clock,
		ExtTimestamp: oldState.ExtTimestamp,
	}
	update = models.GameUpdate{
		OldState: oldState,
		NewState: newState,
	}
	assert.True(t, checker.HasMeaningfulChange(update))

	// Test with period change
	oldState = checker.game.CurrentState
	newState = models.GameState{
		Status:       oldState.Status,
		Home:         oldState.Home,
		Away:         oldState.Away,
		Period:       2,
		Clock:        oldState.Clock,
		ExtTimestamp: oldState.ExtTimestamp,
	}
	update = models.GameUpdate{
		OldState: oldState,
		NewState: newState,
	}
	assert.True(t, checker.HasMeaningfulChange(update))

	// Test with clock change
	oldState = checker.game.CurrentState
	newState = models.GameState{
		Status:       oldState.Status,
		Home:         oldState.Home,
		Away:         oldState.Away,
		Period:       oldState.Period,
		Clock:        "15:30",
		ExtTimestamp: oldState.ExtTimestamp,
	}
	update = models.GameUpdate{
		OldState: oldState,
		NewState: newState,
	}
	assert.True(t, checker.HasMeaningfulChange(update))

	// Test with timestamp change
	oldState = checker.game.CurrentState
	newState = models.GameState{
		Status:       oldState.Status,
		Home:         oldState.Home,
		Away:         oldState.Away,
		Period:       oldState.Period,
		Clock:        oldState.Clock,
		ExtTimestamp: "2024-01-01T12:00:00Z",
	}
	update = models.GameUpdate{
		OldState: oldState,
		NewState: newState,
	}
	assert.True(t, checker.HasMeaningfulChange(update))
}

func TestGameChecker_HandleNoChange(t *testing.T) {
	setupTest(t)

	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")
	memoryStore.AppendActiveGame(game)

	checker, err := NewGameChecker(game.GetGameKey())
	assert.NoError(t, err)

	// Test that HandleNoChange doesn't panic
	assert.NotPanics(t, func() {
		checker.HandleNoChange()
	})
}

func TestGameChecker_HandlePeriodChange(t *testing.T) {
	setupTest(t)

	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")
	game.CurrentState.Period = 2
	memoryStore.AppendActiveGame(game)

	checker, err := NewGameChecker(game.GetGameKey())
	assert.NoError(t, err)

	update := models.GameUpdate{
		OldState: game.CurrentState,
		NewState: game.CurrentState,
	}

	// Test period advancement
	assert.NotPanics(t, func() {
		checker.HandlePeriodChange(1, update)
	})

	// Test no period change
	assert.NotPanics(t, func() {
		checker.HandlePeriodChange(2, update)
	})
}

func TestGameChecker_HandleStatusChange(t *testing.T) {
	setupTest(t)

	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")
	memoryStore.AppendActiveGame(game)

	checker, err := NewGameChecker(game.GetGameKey())
	assert.NoError(t, err)

	// Test status change to active
	checker.game.CurrentState.Status = models.StatusActive
	update := models.GameUpdate{
		OldState: models.GameState{Status: models.StatusUpcoming},
		NewState: checker.game.CurrentState,
	}

	assert.NotPanics(t, func() {
		checker.HandleStatusChange(models.StatusUpcoming, 1, update)
	})

	// Test status change to ended
	checker.game.CurrentState.Status = models.StatusEnded
	update = models.GameUpdate{
		OldState: models.GameState{Status: models.StatusActive},
		NewState: checker.game.CurrentState,
	}

	assert.NotPanics(t, func() {
		checker.HandleStatusChange(models.StatusActive, 1, update)
	})

	// Test no status change
	assert.NotPanics(t, func() {
		checker.HandleStatusChange(models.StatusActive, 1, update)
	})
}

func TestGameChecker_HandleGameEnd(t *testing.T) {
	setupTest(t)

	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")
	memoryStore.AppendActiveGame(game)

	checker, err := NewGameChecker(game.GetGameKey())
	assert.NoError(t, err)

	// Test that HandleGameEnd doesn't panic
	assert.NotPanics(t, func() {
		checker.HandleGameEnd()
	})
}

func TestGameChecker_HandleGameUpdate(t *testing.T) {
	setupTest(t)

	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")
	memoryStore.AppendActiveGame(game)

	checker, err := NewGameChecker(game.GetGameKey())
	assert.NoError(t, err)

	// Test that HandleGameUpdate doesn't panic
	assert.NotPanics(t, func() {
		checker.HandleGameUpdate()
	})
}

func TestCheckGameRefactored(t *testing.T) {
	setupTest(t)

	// Test with non-existent game
	assert.NotPanics(t, func() {
		CheckGameRefactored("non-existent-game")
	})

	// Test with existing game
	game := createTestGame(models.LeagueIdNHL, "WPG", "TOR")
	memoryStore.AppendActiveGame(game)

	assert.NotPanics(t, func() {
		CheckGameRefactored(game.GetGameKey())
	})
}
