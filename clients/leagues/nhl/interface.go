package nhl

type INHLApiClient interface {
	GetNHLSchedule() NHLScheduleResponse
	GetNHLScoreBoard(sGameId string) NHLScoreboardResponse
	GetDiffPatch(gameId string, timestamp string) (NHLDiffPatch, error)
	GetTeam(sLink string) NHLTeamResponse
}
