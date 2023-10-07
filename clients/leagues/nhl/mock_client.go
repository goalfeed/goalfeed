package nhl

import (
	"encoding/json"
)

type MockNHLApiClient struct {
}

var homeScore = 0
var awayScore = 0

func (c MockNHLApiClient) SetHomeScore(score int) {
	homeScore = score
}
func (c MockNHLApiClient) SetAwayScore(score int) {
	awayScore = score
}

func (c MockNHLApiClient) GetNHLSchedule() NHLScheduleResponse {
	var response NHLScheduleResponse
	json.Unmarshal([]byte(UpcomingGamesSchedule), &response)
	return response
}
func (c MockNHLApiClient) GetNHLScoreBoard(sGameId string) NHLScoreboardResponse {
	var response NHLScoreboardResponse
	json.Unmarshal([]byte(ActiveGameScoreboard), &response)
	response.LiveData.Linescore.Teams.Away.Goals = awayScore
	response.LiveData.Linescore.Teams.Home.Goals = homeScore
	return response
}

func (c MockNHLApiClient) GetTeam(sLink string) NHLTeamResponse {
	var response NHLTeamResponse
	json.Unmarshal([]byte(TeamResponseJson), &response)
	return response
}

func (c MockNHLApiClient) GetDiffPatch(gameId string, timestamp string) (NHLDiffPatch, error) {

	var response NHLDiffPatch
	err:= json.Unmarshal([]byte(DiffPatchResponseJson), &response)
	return response, err
}
