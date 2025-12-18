package nhl

type INHLApiClient interface {
	GetNHLSchedule() NHLScheduleResponse
	GetNHLScheduleByDate(date string) NHLScheduleResponse
	GetNHLScoreBoard(sGameId string) NHLScoreboardResponse
	GetTeam(teamAbbr string) NHLTeamResponse
	GetAllTeams() NHLTeamResponse
}
