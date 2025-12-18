package nfl

import (
	"encoding/json"
	"fmt"
	"goalfeed/utils"
)

type NFLAPIClient struct {
}

// fetchByte allows tests to stub the HTTP fetcher
var fetchByte = utils.GetByte

func (c NFLAPIClient) GetNFLSchedule() NFLScheduleResponse {
	var body chan []byte = make(chan []byte)
	url := "https://site.api.espn.com/apis/site/v2/sports/football/nfl/scoreboard"
	go fetchByte(url, body)

	bodyByte := <-body
	var response NFLScheduleResponse
	json.Unmarshal(bodyByte, &response)
	return response
}

func (c NFLAPIClient) GetNFLScheduleByDate(date string) NFLScheduleResponse {
	var body chan []byte = make(chan []byte)
	// Format: YYYYMMDD
	url := fmt.Sprintf("https://site.api.espn.com/apis/site/v2/sports/football/nfl/scoreboard?dates=%s", date)
	go fetchByte(url, body)

	bodyByte := <-body
	var response NFLScheduleResponse
	json.Unmarshal(bodyByte, &response)
	return response
}

func (c NFLAPIClient) GetNFLScoreBoard(gameId string) NFLScoreboardResponse {
	var body chan []byte = make(chan []byte)
	url := fmt.Sprintf("https://site.api.espn.com/apis/site/v2/sports/football/nfl/summary?event=%s", gameId)
	go fetchByte(url, body)

	bodyByte := <-body
	var response NFLScoreboardResponse
	json.Unmarshal(bodyByte, &response)
	return response
}

func (c NFLAPIClient) GetTeam(teamAbbr string) NFLTeamResponse {
	var body chan []byte = make(chan []byte)
	url := fmt.Sprintf("https://site.api.espn.com/apis/site/v2/sports/football/nfl/teams/%s", teamAbbr)
	go fetchByte(url, body)

	bodyByte := <-body
	var response NFLTeamResponse
	json.Unmarshal(bodyByte, &response)
	return response
}

func (c NFLAPIClient) GetAllTeams() NFLTeamResponse {
	var body chan []byte = make(chan []byte)
	url := "https://site.api.espn.com/apis/site/v2/sports/football/nfl/teams"
	go fetchByte(url, body)

	bodyByte := <-body
	var response NFLTeamResponse
	json.Unmarshal(bodyByte, &response)
	return response
}
