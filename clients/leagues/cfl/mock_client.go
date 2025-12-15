package cfl

type MockCFLApiClient struct {
	ScheduleResponse CFLScheduleResponse
	LiveGameResponse CFLLiveGameResponse
}

func (m MockCFLApiClient) GetCFLSchedule() CFLScheduleResponse {
	return m.ScheduleResponse
}

func (m MockCFLApiClient) GetCFLLiveGame(fixtureId string) CFLLiveGameResponse {
	return m.LiveGameResponse
}
