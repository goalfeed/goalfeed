package nfl

type INFLAPIClient interface {
	GetNFLSchedule() NFLScheduleResponse
	GetNFLScheduleByDate(date string) NFLScheduleResponse
	GetNFLScoreBoard(gameId string) NFLScoreboardResponse
	GetTeam(teamAbbr string) NFLTeamResponse
	GetAllTeams() NFLTeamResponse
}
