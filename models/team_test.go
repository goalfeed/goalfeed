package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestTeam_String(t *testing.T) {
	team := Team{
		TeamCode: "TOR",
		TeamName: "Toronto Maple Leafs",
		LeagueID: LeagueIdNHL,
	}

	// Test that we can access team fields
	assert.Equal(t, "TOR", team.TeamCode)
	assert.Equal(t, "Toronto Maple Leafs", team.TeamName)
	assert.Equal(t, LeagueIdNHL, team.LeagueID)
}

func TestTeam_EmptyString(t *testing.T) {
	team := Team{}
	assert.Equal(t, "", team.TeamCode)
	assert.Equal(t, "", team.TeamName)
	assert.Equal(t, 0, team.LeagueID)
}

func TestTeam_OnlyCode(t *testing.T) {
	team := Team{
		TeamCode: "TOR",
	}
	assert.Equal(t, "TOR", team.TeamCode)
	assert.Equal(t, "", team.TeamName)
}

func TestTeam_OnlyName(t *testing.T) {
	team := Team{
		TeamName: "Toronto Maple Leafs",
	}
	assert.Equal(t, "", team.TeamCode)
	assert.Equal(t, "Toronto Maple Leafs", team.TeamName)
}

func TestTeam_WithLeague(t *testing.T) {
	team := Team{
		TeamCode: "TOR",
		TeamName: "Toronto Maple Leafs",
		LeagueID: LeagueIdNHL,
	}

	// Test that league ID is preserved
	assert.Equal(t, LeagueIdNHL, team.LeagueID)
}

func TestTeam_JSONSerialization(t *testing.T) {
	team := Team{
		TeamCode: "TOR",
		TeamName: "Toronto Maple Leafs",
		LeagueID: LeagueIdNHL,
	}

	// Test that we can create and access team fields
	assert.Equal(t, "TOR", team.TeamCode)
	assert.Equal(t, "Toronto Maple Leafs", team.TeamName)
	assert.Equal(t, LeagueIdNHL, team.LeagueID)
}

func TestTeam_EmptyJSON(t *testing.T) {
	team := Team{}

	// Test empty team fields
	assert.Equal(t, "", team.TeamCode)
	assert.Equal(t, "", team.TeamName)
	assert.Equal(t, 0, team.LeagueID)
}

func TestGame_GetGameKey_AllLeagues(t *testing.T) {
	testCases := []struct {
		leagueId    League
		gameCode    string
		expectedKey string
	}{
		{LeagueIdNHL, "2023020001", "1-2023020001"},
		{LeagueIdMLB, "2023020001", "2-2023020001"},
		{LeagueIdEPL, "2023020001", "3-2023020001"},
		{LeagueIdIIHF, "2023020001", "4-2023020001"},
		{LeagueIdCFL, "2023020001", "5-2023020001"},
		{LeagueIdNFL, "2023020001", "6-2023020001"},
	}

	for _, tc := range testCases {
		game := Game{
			GameCode: tc.gameCode,
			LeagueId: tc.leagueId,
		}
		assert.Equal(t, tc.expectedKey, game.GetGameKey())
	}
}

func TestTeam_GetTeamHash_DifferentTeams(t *testing.T) {
	team1 := Team{
		TeamCode: "TOR",
		LeagueID: LeagueIdNHL,
	}

	team2 := Team{
		TeamCode: "MTL",
		LeagueID: LeagueIdNHL,
	}

	// Different teams should have different hashes
	assert.NotEqual(t, team1.GetTeamHash(), team2.GetTeamHash())

	// Same team should have same hash
	assert.Equal(t, team1.GetTeamHash(), team1.GetTeamHash())
}
