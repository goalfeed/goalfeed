package mlb

import (
	"encoding/json"
	"fmt"
	"goalfeed/utils"
	"time"
)

type MLBApiClient struct {
}

// fetchers for testability
var fetchByte = utils.GetByte

func (c MLBApiClient) GetMLBScoreBoard(gameId string) MLBScoreboardResponse {
	var body chan []byte = make(chan []byte)
	url := fmt.Sprintf("https://statsapi.mlb.com/api/v1.1/game/%s/feed/live", gameId)
	go fetchByte(url, body)

	bodyByte := <-body
	var response MLBScoreboardResponse
	json.Unmarshal(bodyByte, &response)
	return response
}

func (c MLBApiClient) GetMLBSchedule() MLBScheduleResponse {
	var body chan []byte = make(chan []byte)

	// Get games for the next 7 days
	now := time.Now()
	startDate := now.Format("2006-01-02")
	endDate := now.AddDate(0, 0, 7).Format("2006-01-02")

	url := fmt.Sprintf("https://statsapi.mlb.com/api/v1/schedule?language=en&sportId=1&startDate=%s&endDate=%s", startDate, endDate)
	go fetchByte(url, body)

	bodyByte := <-body
	var response MLBScheduleResponse
	json.Unmarshal(bodyByte, &response)
	return response
}

func (c MLBApiClient) GetTeam(sLink string) MLBTeamResponse {
	var body chan []byte = make(chan []byte)
	url := fmt.Sprintf("https://statsapi.mlb.com%s", sLink)
	go fetchByte(url, body)

	bodyByte := <-body
	var response MLBTeamResponse
	json.Unmarshal(bodyByte, &response)
	return response
}

func (c MLBApiClient) GetDiffPatch(gameId string, timestamp string) (MLBDiffPatch, error) {
	var body chan []byte = make(chan []byte)
	url := fmt.Sprintf("https://statsapi.mlb.com/api/v1.1/game/%s/feed/live/diffPatch?language=en&startTimecode=%s", gameId, timestamp)
	go fetchByte(url, body)

	bodyByte := <-body
	var response MLBDiffPatch
	err := json.Unmarshal(bodyByte, &response)
	return response, err
}

func (c MLBApiClient) GetAllTeams() MLBTeamResponse {
	var body chan []byte = make(chan []byte)
	url := "https://statsapi.mlb.com/api/v1/teams?sportId=1"
	go fetchByte(url, body)

	bodyByte := <-body
	var response MLBTeamResponse
	json.Unmarshal(bodyByte, &response)
	return response
}
