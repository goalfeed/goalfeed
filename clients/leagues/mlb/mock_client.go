package mlb

import (
	"encoding/json"
)

type MockMLBApiClient struct {
}

var homeScore = 0
var awayScore = 0

func (c MockMLBApiClient) SetHomeScore(score int) {
	homeScore = score
}
func (c MockMLBApiClient) SetAwayScore(score int) {
	awayScore = score
}

func (c MockMLBApiClient) GetMLBSchedule() MLBScheduleResponse {
	var response MLBScheduleResponse
	json.Unmarshal([]byte(UpcomingGamesSchedule), &response)
	return response
}
func (c MockMLBApiClient) GetMLBScoreBoard(sGameId string) MLBScoreboardResponse {
	var response MLBScoreboardResponse
	json.Unmarshal([]byte(ActiveGameScoreboard), &response)
	response.LiveData.Linescore.Teams.Away.Runs = awayScore
	response.LiveData.Linescore.Teams.Home.Runs = homeScore
	return response
}

func (c MockMLBApiClient) GetTeam(sLink string) MLBTeamResponse {
	var response MLBTeamResponse
	json.Unmarshal([]byte(TeamResponseJson), &response)
	return response
}

func (c MockMLBApiClient) GetDiffPatch(gameId string, timestamp string) (MLBDiffPatch, error) {

	var response MLBDiffPatch
	err := json.Unmarshal([]byte(DiffPatchResponseJson), &response)
	return response, err
}
