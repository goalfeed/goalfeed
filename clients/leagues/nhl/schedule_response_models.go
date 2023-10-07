package nhl

import (
	"time"
)

type NHLScheduleTeam struct {
	LeagueRecord struct {
		Wins   int    `json:"wins"`
		Losses int    `json:"losses"`
		Ot     int    `json:"ot"`
		Type   string `json:"type"`
	} `json:"leagueRecord"`
	Score int `json:"score"`
	Team  struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Link string `json:"link"`
	} `json:"team"`
}
type NHLScheduleResponseGame struct {
	GamePk   int    `json:"gamePk"`
	Link     string    `json:"link"`
	GameType string    `json:"gameType"`
	Season   string    `json:"season"`
	GameDate time.Time `json:"gameDate"`
	Status   struct {
		AbstractGameState string `json:"abstractGameState"`
		CodedGameState    string `json:"codedGameState"`
		DetailedState     string `json:"detailedState"`
		StatusCode        string `json:"statusCode"`
		StartTimeTBD      bool   `json:"startTimeTBD"`
	} `json:"status"`
	Teams struct {
		Away NHLScheduleTeam `json:"away"`
		Home NHLScheduleTeam `json:"home"`
	} `json:"teams"`
}
type NHLScheduleResponse struct {
	Copyright    string `json:"copyright"`
	TotalItems   int    `json:"totalItems"`
	TotalEvents  int    `json:"totalEvents"`
	TotalGames   int    `json:"totalGames"`
	TotalMatches int    `json:"totalMatches"`
	Wait         int    `json:"wait"`
	Dates        []struct {
		Date         string                    `json:"date"`
		TotalItems   int                       `json:"totalItems"`
		TotalEvents  int                       `json:"totalEvents"`
		TotalGames   int                       `json:"totalGames"`
		TotalMatches int                       `json:"totalMatches"`
		Games        []NHLScheduleResponseGame `json:"games"`
		Events       []interface{}             `json:"events"`
		Matches      []interface{}             `json:"matches"`
	} `json:"dates"`
}
