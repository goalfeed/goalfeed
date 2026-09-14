package olympics

// MockOlympicsApiClient implements IOlympicsApiClient for tests
type MockOlympicsApiClient struct {
	StartListResp StartListResponse
	ScoresResp    ScoresResponse
}

func (m *MockOlympicsApiClient) GetStartList(genderCode string) (StartListResponse, error) {
	if m.StartListResp != nil {
		return m.StartListResp, nil
	}
	return StartListResponse{}, nil
}

func (m *MockOlympicsApiClient) GetScores(genderCode string) (ScoresResponse, error) {
	return m.ScoresResp, nil
}

func (m *MockOlympicsApiClient) GetScoresForGame(genderCode, phaseCode, eventUnitCode string) (ScoresResponse, error) {
	return m.ScoresResp, nil
}

func (m *MockOlympicsApiClient) GetPlayByPlay(genderCode, phaseCode, eventUnitCode string) (PlayByPlayResponse, error) {
	return PlayByPlayResponse{}, nil
}

func (m *MockOlympicsApiClient) GetScheduleLiteByDay(date string) ([]byte, error) {
	return []byte("{}"), nil
}
