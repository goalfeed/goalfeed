package cfl

type ICFLApiClient interface {
	GetCFLSchedule() CFLScheduleResponse
	GetCFLLiveGame(fixtureId string) CFLLiveGameResponse
}
