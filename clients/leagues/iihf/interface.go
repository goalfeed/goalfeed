package iihf

type IIIHFApiClient interface {
	GetIIHFSchedule(sEventId string) IIHFScheduleResponse
	GetIIHFScoreBoard(sGameId string) IIHFGameScoreResponse
}
