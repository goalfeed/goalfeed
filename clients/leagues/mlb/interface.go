package mlb

type IMLBApiClient interface {
	GetMLBSchedule() MLBScheduleResponse
	GetMLBScoreBoard(sGameId string) MLBScoreboardResponse
	GetDiffPatch(gameId string, timestamp string) (MLBDiffPatch, error)
	GetTeam(sLink string) MLBTeamResponse
	GetAllTeams() MLBTeamResponse
}
