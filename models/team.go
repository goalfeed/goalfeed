package models

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

type Team struct {
	ID       int    `json:"TeamId"`
	TeamCode string `json:"TeamCode"`
	TeamName string `json:"TeamName"`
	LeagueID int    `json:"LeagueID"`
	ExtID    string `json:"ExtID"`
}

// GetTeamHash generates a unique has for the team based on the TeamCode and LeagueId
// I don't know why I made this field illegible to humans when I originally did it, but
// we need to continue to include it in case it is in use.
func (t Team) GetTeamHash() string {
	data := []byte(fmt.Sprintf("%s%d", t.TeamCode, t.LeagueID))
	hash := md5.Sum(data)

	return hex.EncodeToString(hash[:])
}
