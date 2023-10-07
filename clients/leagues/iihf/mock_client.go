package iihf

import (
	"encoding/json"
)

type MockIIHFApiClient struct {
}

var homeScore = 0
var awayScore = 0

func (c MockIIHFApiClient) SetHomeScore(score int) {
	homeScore = score
}
func (c MockIIHFApiClient) SetAwayScore(score int) {
	awayScore = score
}

func (c MockIIHFApiClient) GetIIHFSchedule(sEventId string) IIHFScheduleResponse {
	var response IIHFScheduleResponse
	json.Unmarshal([]byte(UpcomingGamesSchedule), &response)
	return response
}
func (c MockIIHFApiClient) GetIIHFScoreBoard(sGameId string) IIHFGameScoreResponse {
	var response IIHFGameScoreResponse
	json.Unmarshal([]byte(ActiveGameScoreboard), &response)
	response.CurrentScore.Away = awayScore
	response.CurrentScore.Home = homeScore
	return response
}
