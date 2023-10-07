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
	url := fmt.Sprintf("https://statsapi.web.nhl.com/api/v1/game/%s/feed/live", gameId)
	go utils.GetByte(url, body)

	bodyByte := <-body
	var response NHLScoreboardResponse
	// fmt.Println(string(bodyByte))
	json.Unmarshal(bodyByte, &response)
	return response
}

func (c NHLApiClient) GetNHLSchedule() NHLScheduleResponse {
	var body chan []byte = make(chan []byte)
	url := "https://statsapi.web.nhl.com/api/v1/schedule"
	go utils.GetByte(url, body)

	bodyByte := <-body
	var response NHLScheduleResponse
	json.Unmarshal(bodyByte, &response)
	return response
}

func (c NHLApiClient) GetTeam(sLink string) NHLTeamResponse {
	var body chan []byte = make(chan []byte)
	url := fmt.Sprintf("https://statsapi.web.nhl.com%s", sLink)
	go utils.GetByte(url, body)

	bodyByte := <-body
	var response NHLTeamResponse
	json.Unmarshal(bodyByte, &response)
	return response
}

func (c NHLApiClient) GetDiffPatch(gameId string, timestamp string) (NHLDiffPatch, error) {
	var body chan []byte = make(chan []byte)
	url := fmt.Sprintf("https://statsapi.web.nhl.com/api/v1/game/%s/feed/live/diffPatch?site=en_nhl&startTimecode=%s", gameId, timestamp)
	go utils.GetByte(url, body)

	bodyByte := <-body
	var response NHLDiffPatch
	err := json.Unmarshal(bodyByte, &response)
	return response, err
}
