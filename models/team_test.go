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
