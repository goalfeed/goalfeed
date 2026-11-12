package nhl

import (
	"encoding/json"
	"strconv"
)

type MockNHLApiClient struct {
	mockedGameStatus      string
	GetNHLScheduleCalls   int
	GetNHLScoreBoardCalls int
	// RightRail is returned by GetNHLRightRail; nil means an empty panel.
	RightRail    *NHLRightRailResponse
	RightRailErr error
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

func (c MockNHLApiClient) GetNHLScheduleByDate(date string) NHLScheduleResponse {
	var response NHLScheduleResponse
	c.GetNHLScheduleCalls++
	json.Unmarshal([]byte(UpcomingGamesSchedule), &response)
	return response
}
func (c MockNHLApiClient) GetNHLScoreBoard(sGameId string) NHLScoreboardResponse {
	var response NHLScoreboardResponse
	c.GetNHLScoreBoardCalls++
	json.Unmarshal([]byte(ActiveGameScoreboard), &response)
	// The canned payload has no top-level id; stand in for the landing of
	// whichever game was asked for, as the real endpoint does.
	if id, err := strconv.Atoi(sGameId); err == nil {
		response.ID = id
	}
	response.AwayTeam.Score = awayScore
	response.HomeTeam.Score = homeScore
	return response
}

func (c MockNHLApiClient) GetTeam(sLink string) NHLTeamResponse {
	var response NHLTeamResponse
	json.Unmarshal([]byte(TeamResponseJson), &response)
	return response
}

func (c MockNHLApiClient) GetAllTeams() NHLTeamResponse {
	var response NHLTeamResponse
	json.Unmarshal([]byte(TeamResponseJson), &response)
	return response
}

func (c MockNHLApiClient) GetNHLRightRail(sGameId string) (NHLRightRailResponse, error) {
	if c.RightRail == nil {
		return NHLRightRailResponse{}, c.RightRailErr
	}
	return *c.RightRail, c.RightRailErr
}
