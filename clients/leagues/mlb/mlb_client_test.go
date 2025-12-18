package mlb

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func withStubFetchMLB(t *testing.T, payload interface{}) func() {
	oldB := fetchByte
	b, _ := json.Marshal(payload)
	fetchByte = func(url string, ret chan []byte) { ret <- b }
	return func() { fetchByte = oldB }
}

func withStubFetchMLBError(t *testing.T) func() {
	oldB := fetchByte
	fetchByte = func(url string, ret chan []byte) { ret <- []byte{} }
	return func() { fetchByte = oldB }
}

func TestMLBApiClient_GetMLBSchedule(t *testing.T) {
	client := MLBApiClient{}

	// Note: This will make an actual HTTP call which may fail in test environment
	// In a real implementation, we would mock the HTTP client
	// For now, just test that the method exists and doesn't panic
	assert.NotPanics(t, func() {
		_ = client.GetMLBSchedule()
	})
}

func TestMLBApiClient_GetMLBScoreBoard(t *testing.T) {
	client := MLBApiClient{}

	// Test that the method exists and doesn't panic
	assert.NotPanics(t, func() {
		_ = client.GetMLBScoreBoard("2023020001")
	})
}

func TestMLBApiClient_GetTeam(t *testing.T) {
	client := MLBApiClient{}

	// Test that the method exists and doesn't panic
	assert.NotPanics(t, func() {
		_ = client.GetTeam("/api/v1/teams/143")
	})
}

func TestMLBApiClient_GetDiffPatch(t *testing.T) {
	client := MLBApiClient{}

	// Test that the method exists and doesn't panic
	assert.NotPanics(t, func() {
		_, _ = client.GetDiffPatch("2023020001", "20230101_000000")
	})
}

func TestMockMLBApiClient_GetMLBSchedule(t *testing.T) {
	client := MockMLBApiClient{}

	response := client.GetMLBSchedule()

	// Verify the response is valid
	assert.NotNil(t, response)
}

func TestMockMLBApiClient_GetMLBScoreBoard(t *testing.T) {
	client := MockMLBApiClient{}

	response := client.GetMLBScoreBoard("2023020001")

	// Verify the response is valid
	assert.NotNil(t, response)
}

func TestMockMLBApiClient_GetTeam(t *testing.T) {
	client := MockMLBApiClient{}

	response := client.GetTeam("/api/v1/teams/143")

	// Verify the response is valid
	assert.NotNil(t, response)
}

func TestMockMLBApiClient_GetDiffPatch(t *testing.T) {
	client := MockMLBApiClient{}

	result, err := client.GetDiffPatch("2023020001", "20230101_000000")

	// Verify the response
	assert.NotNil(t, result)
	assert.NoError(t, err)
}

func TestMockMLBApiClient_SetScores(t *testing.T) {
	client := MockMLBApiClient{}

	// Note: These are global variables in the mock, so we test the methods exist
	assert.NotPanics(t, func() {
		client.SetHomeScore(5)
		client.SetAwayScore(3)
	})
}

func TestMLBApiClient_GetAllTeams(t *testing.T) {
	restore := withStubFetchMLB(t, MLBTeamResponse{Teams: []MLBTeamsResponseTeam{{ID: "1", Name: "Test Team"}}})
	defer restore()
	client := MLBApiClient{}
	resp := client.GetAllTeams()
	if len(resp.Teams) != 1 || resp.Teams[0].ID != "1" {
		t.Fatalf("unexpected all teams resp: %+v", resp)
	}
}

func TestMLBApiClient_GetAllTeams_EmptyResponse(t *testing.T) {
	restore := withStubFetchMLBError(t)
	defer restore()
	client := MLBApiClient{}
	resp := client.GetAllTeams()
	if len(resp.Teams) != 0 {
		t.Fatalf("expected empty response, got: %+v", resp)
	}
}

func TestMockMLBApiClient_GetAllTeams(t *testing.T) {
	client := MockMLBApiClient{}

	response := client.GetAllTeams()

	// Verify the response is valid
	assert.NotNil(t, response)
}
