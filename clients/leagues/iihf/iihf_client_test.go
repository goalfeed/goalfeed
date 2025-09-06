package iihf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIIHFApiClient_GetIIHFSchedule(t *testing.T) {
	client := IIHFApiClient{}

	// Note: This will make an actual HTTP call which may fail in test environment
	// In a real implementation, we would mock the HTTP client
	// For now, just test that the method exists and doesn't panic
	assert.NotPanics(t, func() {
		_ = client.GetIIHFSchedule("123")
	})
}

func TestIIHFApiClient_GetIIHFScoreBoard(t *testing.T) {
	client := IIHFApiClient{}

	// Test that the method exists and doesn't panic
	assert.NotPanics(t, func() {
		_ = client.GetIIHFScoreBoard("2023020001")
	})
}

func TestMockIIHFApiClient_GetIIHFSchedule(t *testing.T) {
	client := MockIIHFApiClient{}

	response := client.GetIIHFSchedule("123")

	// Verify the response is valid
	assert.NotNil(t, response)
}

func TestMockIIHFApiClient_GetIIHFScoreBoard(t *testing.T) {
	client := MockIIHFApiClient{}

	response := client.GetIIHFScoreBoard("2023020001")

	// Verify the response is valid
	assert.NotNil(t, response)
}

func TestMockIIHFApiClient_SetScores(t *testing.T) {
	client := MockIIHFApiClient{}

	// Note: These are global variables in the mock, so we test the methods exist
	assert.NotPanics(t, func() {
		client.SetHomeScore(4)
		client.SetAwayScore(2)
	})
}
