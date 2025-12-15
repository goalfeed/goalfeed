package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventPriorityIconColor(t *testing.T) {
	events := []struct {
		etype    EventType
		priority EventPriority
		icon     string
		color    string
	}{
		{EventTypeGoal, PriorityHigh, "🏒", "green"},
		{EventTypeTouchdown, PriorityHigh, "🏈", "green"},
		{EventTypeHomeRun, PriorityHigh, "⚾", "green"},
		{EventTypePenalty, PriorityHigh, "⚠️", "red"},
		{EventTypePowerPlay, PriorityNormal, "⚡", "yellow"},
		{EventTypeShot, PriorityNormal, "🎯", "gray"},
		{EventTypeSave, PriorityNormal, "🛡️", "gray"},
		{EventTypeStrikeout, PriorityNormal, "⚡", "yellow"},
		{EventTypeWalk, PriorityNormal, "🚶", "gray"},
		{EventTypeError, PriorityNormal, "❌", "red"},
		{EventTypeGameStart, PriorityNormal, "🏁", "blue"},
		{EventTypeGameEnd, PriorityNormal, "🏁", "blue"},
		{EventTypePeriodStart, PriorityNormal, "⏰", "purple"},
		{EventTypePeriodEnd, PriorityNormal, "⏰", "purple"},
	}

	for _, tc := range events {
		e := Event{Type: tc.etype}
		assert.Equal(t, tc.priority, e.GetEventPriority(), "priority for %s", tc.etype)
		assert.Equal(t, tc.icon, e.GetEventIcon(), "icon for %s", tc.etype)
		assert.Equal(t, tc.color, e.GetEventColor(), "color for %s", tc.etype)
	}
}

func TestGameStatusJSON(t *testing.T) {
	// Marshal
	data, err := json.Marshal(GameStatus(StatusActive))
	assert.NoError(t, err)
	assert.Equal(t, "\"active\"", string(data))

	// Unmarshal from string
	var gs GameStatus
	assert.NoError(t, json.Unmarshal([]byte("\"upcoming\""), &gs))
	assert.Equal(t, GameStatus(StatusUpcoming), gs)
	assert.NoError(t, json.Unmarshal([]byte("\"active\""), &gs))
	assert.Equal(t, GameStatus(StatusActive), gs)
	assert.NoError(t, json.Unmarshal([]byte("\"delayed\""), &gs))
	assert.Equal(t, GameStatus(StatusDelayed), gs)
	assert.NoError(t, json.Unmarshal([]byte("\"ended\""), &gs))
	assert.Equal(t, GameStatus(StatusEnded), gs)

	// Unknown string defaults to upcoming
	assert.NoError(t, json.Unmarshal([]byte("\"something\""), &gs))
	assert.Equal(t, GameStatus(StatusUpcoming), gs)

	// Unmarshal from number
	assert.NoError(t, json.Unmarshal([]byte("2"), &gs))
	assert.Equal(t, GameStatus(2), gs)

	// Marshal unknown should produce "unknown"
	b, err := json.Marshal(GameStatus(999))
	assert.NoError(t, err)
	assert.Equal(t, "\"unknown\"", string(b))
}

func TestGameStatusMarshalJSON_AllCases(t *testing.T) {
	// Test all GameStatus values for MarshalJSON
	testCases := []struct {
		status   GameStatus
		expected string
	}{
		{StatusUpcoming, "\"upcoming\""},
		{StatusActive, "\"active\""},
		{StatusDelayed, "\"delayed\""},
		{StatusEnded, "\"ended\""},
		{GameStatus(999), "\"unknown\""}, // Default case
	}

	for _, tc := range testCases {
		data, err := json.Marshal(tc.status)
		assert.NoError(t, err)
		assert.Equal(t, tc.expected, string(data))
	}
}
