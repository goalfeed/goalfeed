package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetTeamHash(t *testing.T) {
	team := Team{
		TeamCode: "WPG",
		LeagueID: LeagueIdNHL,
	}
	assert.Equal(t, "2403d3bd4d1bc4befe4f79d15ca37749", team.GetTeamHash())
}

func TestGetGameKey(t *testing.T) {
	game := Game{
		GameCode: "2023020001",
		LeagueId: LeagueIdNHL,
	}

	expectedKey := "1-2023020001"
	assert.Equal(t, expectedKey, game.GetGameKey())

	// Test with MLB league
	game.LeagueId = LeagueIdMLB
	expectedKey = "2-2023020001"
	assert.Equal(t, expectedKey, game.GetGameKey())
}
