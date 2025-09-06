package cfl

import (
	"encoding/json"
	"fmt"
	"goalfeed/utils"
)

type CFLApiClient struct {
}

func (c CFLApiClient) GetCFLSchedule() CFLScheduleResponse {
	var body chan []byte = make(chan []byte)
	url := "https://cflscoreboard.cfl.ca/json/scoreboard/rounds.json"
	go utils.GetByte(url, body)

	bodyByte := <-body
	var response CFLScheduleResponse
	json.Unmarshal(bodyByte, &response)
	return response
}

func (c CFLApiClient) GetCFLLiveGame(fixtureId string) CFLLiveGameResponse {
	var body chan []byte = make(chan []byte)
	url := fmt.Sprintf("https://gsm-widgets.betstream.betgenius.com/widget-data/multisportgametracker?productName=democfl_light&region=MB&country=CA&fixtureId=%s&activeContent=court&sport=AmericanFootball&sportId=17&culture[0]=en-US&competitionId=1035&isUsingBetGeniusId=true", fixtureId)
	go utils.GetByte(url, body)

	bodyByte := <-body
	var response CFLLiveGameResponse
	json.Unmarshal(bodyByte, &response)
	return response
}
