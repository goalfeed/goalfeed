package nfl

type INFLAPIClient interface {
	GetNFLSchedule() NFLScheduleResponse
	GetNFLScoreBoard(gameId string) NFLScoreboardResponse
	GetTeam(teamAbbr string) NFLTeamResponse
	GetAllTeams() NFLTeamResponse
}
