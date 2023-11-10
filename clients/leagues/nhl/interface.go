package nhl

type INHLApiClient interface {
	GetNHLSchedule() NHLScheduleResponse
	GetNHLScoreBoard(sGameId string) NHLScoreboardResponse
	GetTeam(teamAbbr string) NHLTeamResponse
}
