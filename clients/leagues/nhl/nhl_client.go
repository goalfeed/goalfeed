package nhl

import (
	"encoding/json"
	"fmt"
	"goalfeed/utils"
)

type NHLApiClient struct {
}

func (c NHLApiClient) GetNHLScoreBoard(gameId string) NHLScoreboardResponse {
	var body chan []byte = make(chan []byte)
	url := fmt.Sprintf("https://api-web.nhle.com/v1/gamecenter/%s/landing", gameId) // Updated URL
	go utils.GetByte(url, body)

	bodyByte := <-body
	var response NHLScoreboardResponse
	json.Unmarshal(bodyByte, &response)
	return response
}

func (c NHLApiClient) GetNHLSchedule() NHLScheduleResponse {
	var body chan []byte = make(chan []byte)
	url := "https://api-web.nhle.com/v1/schedule/now" // Updated URL
	go utils.GetByte(url, body)

	bodyByte := <-body
	var response NHLScheduleResponse
	json.Unmarshal(bodyByte, &response)
	return response
}

func (c NHLApiClient) GetTeam(teamAbbr string) NHLTeamResponse {
	var body chan []byte = make(chan []byte)
	url := fmt.Sprintf("https://api-web.nhle.com/v1/club-schedule-season/%s/now", teamAbbr) // Updated URL
	go utils.GetByte(url, body)

	bodyByte := <-body
	var response NHLTeamResponse
	json.Unmarshal(bodyByte, &response)
	return response
}

func (c NHLApiClient) GetDiffPatch(gameId string, timestamp string) (NHLDiffPatch, error) {
	var body chan []byte = make(chan []byte)
	url := fmt.Sprintf("https://api-web.nhle.com/v1/game/%s/feed/live/diffPatch?site=en_nhl&startTimecode=%s", gameId, timestamp) // Updated URL
	go utils.GetByte(url, body)

	bodyByte := <-body
	var response NHLDiffPatch
	err := json.Unmarshal(bodyByte, &response)
	return response, err
}

func (c NHLApiClient) GetAllTeams() NHLTeamResponse {
	var body chan []byte = make(chan []byte)
	url := "https://api-web.nhle.com/v1/teams"
	go utils.GetByte(url, body)

	bodyByte := <-body
	var response NHLTeamResponse
	json.Unmarshal(bodyByte, &response)
	return response
}
