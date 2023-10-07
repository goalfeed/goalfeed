package iihf

import (
	"encoding/json"
	"fmt"
	"goalfeed/utils"
)

type IIHFApiClient struct {
}

func (c IIHFApiClient) GetIIHFScoreBoard(sGameId string) IIHFGameScoreResponse {
	var body chan []byte = make(chan []byte)
	url := fmt.Sprintf("https://realtime.iihf.com/gamestate/GetLatestState/%s", sGameId)
	go utils.GetByte(url, body)

	bodyByte := <-body
	var response IIHFGameScoreResponse
	// fmt.Println(string(bodyByte))
	json.Unmarshal(bodyByte, &response)
	return response
}

func (c IIHFApiClient) GetIIHFSchedule(sEventId string) IIHFScheduleResponse {
	var body chan []byte = make(chan []byte)
	url := fmt.Sprintf("https://realtime.iihf.com/gamestate/GetLatestScoresState/%s", sEventId)
	go utils.GetByte(url, body)

	bodyByte := <-body
	var response IIHFScheduleResponse
	// fmt.Println(string(bodyByte))
	json.Unmarshal(bodyByte, &response)
	return response
}

