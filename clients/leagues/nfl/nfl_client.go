package nfl

import (
	"encoding/json"
	"fmt"
	"goalfeed/utils"
)

type NFLAPIClient struct {
}

func (c NFLAPIClient) GetNFLSchedule() NFLScheduleResponse {
	var body chan []byte = make(chan []byte)
	url := "https://site.api.espn.com/apis/site/v2/sports/football/nfl/scoreboard"
	go utils.GetByte(url, body)

	bodyByte := <-body
	var response NFLScheduleResponse
	json.Unmarshal(bodyByte, &response)
	return response
}

func (c NFLAPIClient) GetNFLScoreBoard(gameId string) NFLScoreboardResponse {
	var body chan []byte = make(chan []byte)
	url := fmt.Sprintf("https://site.api.espn.com/apis/site/v2/sports/football/nfl/summary?event=%s", gameId)
	go utils.GetByte(url, body)

	bodyByte := <-body
	var response NFLScoreboardResponse
	json.Unmarshal(bodyByte, &response)
	return response
}

func (c NFLAPIClient) GetTeam(teamAbbr string) NFLTeamResponse {
	var body chan []byte = make(chan []byte)
	url := fmt.Sprintf("https://site.api.espn.com/apis/site/v2/sports/football/nfl/teams/%s", teamAbbr)
	go utils.GetByte(url, body)

	bodyByte := <-body
	var response NFLTeamResponse
	json.Unmarshal(bodyByte, &response)
	return response
}

func (c NFLAPIClient) GetAllTeams() NFLTeamResponse {
	var body chan []byte = make(chan []byte)
	url := "https://site.api.espn.com/apis/site/v2/sports/football/nfl/teams"
	go utils.GetByte(url, body)

	bodyByte := <-body
	var response NFLTeamResponse
	json.Unmarshal(bodyByte, &response)
	return response
}
