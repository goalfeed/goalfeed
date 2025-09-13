package cfl

import (
	"encoding/json"
	"fmt"
	"goalfeed/utils"
)

var logger = utils.GetLogger()

type CFLApiClient struct {
}

// fetchers for testability
var fetchByte = utils.GetByte
var fetchByteWithHeaders = utils.GetByteWithHeaders

func (c CFLApiClient) GetCFLSchedule() CFLScheduleResponse {
	var body chan []byte = make(chan []byte)
	url := "https://cflscoreboard.cfl.ca/json/scoreboard/rounds.json"
	// Primary attempt: headers; omit Accept-Encoding so Go auto-decompresses gzip
	go fetchByteWithHeaders(url, body, map[string]string{
		"Accept":     "application/json",
		"User-Agent": "Goalfeed/1.0",
	})

	bodyByte := <-body
	var response CFLScheduleResponse

	// If empty, fall back to no-headers fetch immediately
	if len(bodyByte) == 0 {
		logger.Warn("CFL schedule fetch returned empty body; retrying without headers")
		retry := make(chan []byte)
		go fetchByte(url, retry)
		bodyByte = <-retry
	}

	if len(bodyByte) == 0 {
		return CFLScheduleResponse{}
	}

	// Try to unmarshal the JSON
	if err := json.Unmarshal(bodyByte, &response); err != nil || len(response) == 0 {
		if err != nil {
			logger.Warnf("CFL schedule JSON unmarshal failed on primary: %v; retrying without headers", err)
		} else {
			logger.Warn("CFL schedule parsed but empty; retrying without headers")
		}
		// Retry without headers
		retry := make(chan []byte)
		go fetchByte(url, retry)
		retryBody := <-retry
		if len(retryBody) == 0 {
			return CFLScheduleResponse{}
		}
		var retryResp CFLScheduleResponse
		if err2 := json.Unmarshal(retryBody, &retryResp); err2 == nil {
			return retryResp
		}
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
