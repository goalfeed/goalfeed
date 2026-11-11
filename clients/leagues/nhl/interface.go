package nhl

type INHLApiClient interface {
	GetNHLSchedule() NHLScheduleResponse
	GetNHLScheduleByDate(date string) NHLScheduleResponse
	GetNHLScoreBoard(sGameId string) NHLScoreboardResponse
	GetNHLRightRail(sGameId string) (NHLRightRailResponse, error)
	GetTeam(teamAbbr string) NHLTeamResponse
	GetAllTeams() NHLTeamResponse
}
