package models

type Event struct {
	TeamCode     string `json:"team"`
	TeamName     string `json:"team_name"`
	TeamHash     string `json:"team_hash"`
	LeagueId     int    `json:"league_id"`
	LeagueName   string `json:"league_name"`
	OpponentCode string `json:"opponent"`
	OpponentName string `json:"opponent_name"`
	OpponentHash string `json:"opponent_hash"`
}
