package cfl

import (
	"encoding/json"
	"fmt"
	"goalfeed/utils"
)

type CFLApiClient struct {
}

// fetchers for testability
var fetchByte = utils.GetByte
var fetchByteWithHeaders = utils.GetByteWithHeaders

func (c CFLApiClient) GetCFLSchedule() CFLScheduleResponse {
	var body chan []byte = make(chan []byte)
	url := "https://cflscoreboard.cfl.ca/json/scoreboard/rounds.json"
	go fetchByteWithHeaders(url, body, map[string]string{
		"Accept-Encoding": "identity",
		"User-Agent":      "Goalfeed/1.0",
		"Accept":          "application/json",
	})

	bodyByte := <-body
	var response CFLScheduleResponse

	// Check if the response is valid JSON
	if len(bodyByte) == 0 {
		// Return empty response if no data
		return CFLScheduleResponse{}
	}

	// Try to unmarshal the JSON
	if err := json.Unmarshal(bodyByte, &response); err != nil {
		// If unmarshaling fails, return empty response
		// This handles cases where the API returns compressed or corrupted data
		return CFLScheduleResponse{}
	}

	return response
}

func (c CFLApiClient) GetCFLLiveGame(fixtureId string) CFLLiveGameResponse {
	var body chan []byte = make(chan []byte)
	url := fmt.Sprintf("https://gsm-widgets.betstream.betgenius.com/widget-data/multisportgametracker?productName=democfl_light&region=MB&country=CA&fixtureId=%s&activeContent=court&sport=AmericanFootball&sportId=17&culture[0]=en-US&competitionId=1035&isUsingBetGeniusId=true", fixtureId)
	go fetchByte(url, body)

	bodyByte := <-body
	var response CFLLiveGameResponse

	// Check if the response is valid JSON
	if len(bodyByte) == 0 {
		// Return empty response if no data
		return CFLLiveGameResponse{}
	}

	// Try to unmarshal the JSON
	if err := json.Unmarshal(bodyByte, &response); err != nil {
		// If unmarshaling fails, return empty response
		// This handles cases where the API returns compressed or corrupted data
		return CFLLiveGameResponse{}
	}

	return response
}
