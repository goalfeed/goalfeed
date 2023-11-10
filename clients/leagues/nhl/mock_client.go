package nhl

import (
	"encoding/json"
)

type MockNHLApiClient struct {
	mockedGameStatus      string
	GetNHLScheduleCalls   int
	GetNHLScoreBoardCalls int
}

var homeScore = 0
var awayScore = 0

func (m *MockNHLApiClient) SetGameStatus(status string) {
	m.mockedGameStatus = status
}
func (c MockNHLApiClient) SetHomeScore(score int) {
	homeScore = score
}
func (c MockNHLApiClient) SetAwayScore(score int) {
	awayScore = score
}

func (c MockNHLApiClient) GetNHLSchedule() NHLScheduleResponse {
	var response NHLScheduleResponse
	c.GetNHLScheduleCalls++
	json.Unmarshal([]byte(UpcomingGamesSchedule), &response)
	return response
}
func (c MockNHLApiClient) GetNHLScoreBoard(sGameId string) NHLScoreboardResponse {
	var response NHLScoreboardResponse
	c.GetNHLScoreBoardCalls++
	json.Unmarshal([]byte(ActiveGameScoreboard), &response)
	response.AwayTeam.Score = awayScore
	response.HomeTeam.Score = homeScore
	return response
}

func (c MockNHLApiClient) GetTeam(sLink string) NHLTeamResponse {
	var response NHLTeamResponse
	json.Unmarshal([]byte(TeamResponseJson), &response)
	return response
}

func (c MockNHLApiClient) GetDiffPatch(gameId string, timestamp string) (NHLDiffPatch, error) {

	var response NHLDiffPatch
	err := json.Unmarshal([]byte(DiffPatchResponseJson), &response)
	return response, err
}
