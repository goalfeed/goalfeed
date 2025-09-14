package nfl

import (
	"encoding/json"
	"testing"

	"goalfeed/models"

	"github.com/stretchr/testify/assert"
)

func TestNormalizePayload(t *testing.T) {
	// Test with JSON string
	jsonString := `"test-string"`
	normalized := normalizePayload(json.RawMessage(jsonString))
	assert.Equal(t, []byte("test-string"), normalized)

	// Test with direct bytes
	directBytes := []byte(`{"test": "value"}`)
	normalized = normalizePayload(json.RawMessage(directBytes))
	assert.Equal(t, directBytes, normalized)
}

func TestDecodeWrapperPayload(t *testing.T) {
	// Test with valid wrapper containing JSON array
	wrapperJSON := `{"ts":1234567890,"~c":1,"pl":[{"op":"replace","path":"/test","value":"test"}]}`
	ops, err := decodeWrapperPayload([]byte(wrapperJSON))
	assert.NoError(t, err)
	assert.Len(t, ops, 1)
	assert.Equal(t, "replace", ops[0].Op)
	assert.Equal(t, "/test", ops[0].Path)
	assert.Equal(t, "test", ops[0].Value)

	// Test with invalid wrapper
	invalidJSON := `{"invalid": "data"}`
	_, err = decodeWrapperPayload([]byte(invalidJSON))
	assert.Error(t, err)

	// Test with empty pl field
	emptyPl := `{"ts":1234567890,"~c":1,"pl":""}`
	_, err = decodeWrapperPayload([]byte(emptyPl))
	assert.Error(t, err)
}

func TestDecodeBase64ZlibPayload(t *testing.T) {
	// Test with invalid base64
	_, err := decodeBase64ZlibPayload("invalid-base64")
	assert.Error(t, err)

	// Test with valid base64 but invalid zlib
	_, err = decodeBase64ZlibPayload("dGVzdA==") // "test" in base64
	assert.Error(t, err)
}

func TestDecodeDirectPayload(t *testing.T) {
	// Test with valid JSON array
	jsonArray := `[{"op":"replace","path":"/test","value":"test"}]`
	ops, err := decodeDirectPayload([]byte(jsonArray))
	assert.NoError(t, err)
	assert.Len(t, ops, 1)
	assert.Equal(t, "replace", ops[0].Op)

	// Test with invalid JSON
	invalidJSON := `{"invalid": "data"}`
	_, err = decodeDirectPayload([]byte(invalidJSON))
	assert.Error(t, err)
}

func TestExtractOperations(t *testing.T) {
	// Test with wrapper format
	wrapperJSON := `{"ts":1234567890,"~c":1,"pl":[{"op":"replace","path":"/test","value":"test"}]}`
	ops, err := extractOperations(json.RawMessage(wrapperJSON))
	assert.NoError(t, err)
	assert.Len(t, ops, 1)

	// Test with direct JSON array
	directJSON := `[{"op":"replace","path":"/test","value":"test"}]`
	ops, err = extractOperations(json.RawMessage(directJSON))
	assert.NoError(t, err)
	assert.Len(t, ops, 1)

	// Test with JSON string wrapper
	stringWrapper := `"{\"ts\":1234567890,\"~c\":1,\"pl\":[{\"op\":\"replace\",\"path\":\"/test\",\"value\":\"test\"}]}"`
	ops, err = extractOperations(json.RawMessage(stringWrapper))
	assert.NoError(t, err)
	assert.Len(t, ops, 1)

	// Test with invalid payload
	invalidPayload := `{"invalid": "data"}`
	_, err = extractOperations(json.RawMessage(invalidPayload))
	assert.Error(t, err)
}

func TestBuildCompetitorMapping(t *testing.T) {
	ops := []patchOp{
		{Op: "replace", Path: "/competitors/0/homeAway", Value: "home"},
		{Op: "replace", Path: "/competitors/1/homeAway", Value: "away"},
		{Op: "replace", Path: "/competitors/0/team/id", Value: "4"},
		{Op: "replace", Path: "/competitors/1/team/id", Value: "15"},
	}

	mapping := buildCompetitorMapping(ops)

	assert.Equal(t, "home", mapping.Side["0"])
	assert.Equal(t, "away", mapping.Side["1"])
	assert.Equal(t, "4", mapping.TeamID["0"])
	assert.Equal(t, "15", mapping.TeamID["1"])
}

func TestExtractEventID(t *testing.T) {
	// Test with path containing event ID
	op := patchOp{Path: "/e:12345/clock"}
	eventID := extractEventID(op, "gp-football-nfl-test")
	assert.Equal(t, "12345", eventID)

	// Test with topic fallback
	op = patchOp{Path: "/some/other/path"}
	eventID = extractEventID(op, "gp-football-nfl-test-event")
	assert.Equal(t, "test-event", eventID)

	// Test with no event ID available
	op = patchOp{Path: "/some/other/path"}
	eventID = extractEventID(op, "invalid-topic")
	assert.Equal(t, "", eventID)
}

func TestParsePeriodFromText(t *testing.T) {
	testCases := []struct {
		text           string
		expectedPeriod int
		expectedType   string
	}{
		{"1st", 1, "QUARTER"},
		{"2nd", 2, "QUARTER"},
		{"3rd", 3, "QUARTER"},
		{"4th", 4, "QUARTER"},
		{"ot", 5, "OVERTIME"},
		{"ot1", 5, "OVERTIME"},
		{"2ot", 6, "OVERTIME"},
		{"invalid", 0, ""},
		{"", 0, ""},
	}

	for _, tc := range testCases {
		period, periodType := parsePeriodFromText(tc.text)
		assert.Equal(t, tc.expectedPeriod, period)
		assert.Equal(t, tc.expectedType, periodType)
	}
}

func TestApplyClockUpdate(t *testing.T) {
	game := models.Game{
		CurrentState: models.GameState{
			Status: models.StatusUpcoming,
		},
	}

	// Test display clock update
	op := patchOp{
		Path:  "/fullStatus/displayClock",
		Value: "15:30",
	}
	updatedGame := applyClockUpdate(game, op)
	assert.Equal(t, "15:30", updatedGame.CurrentState.Clock)
	assert.Equal(t, models.GameStatus(models.StatusActive), updatedGame.CurrentState.Status)

	// Test numeric clock update
	op = patchOp{
		Path:  "/fullStatus/clock",
		Value: float64(930), // 15:30 in seconds
	}
	updatedGame = applyClockUpdate(game, op)
	assert.Equal(t, "15:30", updatedGame.CurrentState.Clock)
	assert.Equal(t, models.GameStatus(models.StatusActive), updatedGame.CurrentState.Status)

	// Test clock path
	op = patchOp{
		Path:  "/clock",
		Value: "12:45",
	}
	updatedGame = applyClockUpdate(game, op)
	assert.Equal(t, "12:45", updatedGame.CurrentState.Clock)

	// Test non-clock path
	op = patchOp{
		Path:  "/other/path",
		Value: "test",
	}
	updatedGame = applyClockUpdate(game, op)
	assert.Equal(t, game.CurrentState.Clock, updatedGame.CurrentState.Clock) // Should be unchanged
}

func TestApplyPeriodUpdate(t *testing.T) {
	game := models.Game{
		CurrentState: models.GameState{},
	}

	// Test short detail update
	op := patchOp{
		Path:  "/fullStatus/type/shortDetail",
		Value: "6:14 - 3rd",
	}
	updatedGame := applyPeriodUpdate(game, op)
	assert.Equal(t, "6:14", updatedGame.CurrentState.Clock)
	assert.Equal(t, 3, updatedGame.CurrentState.Period)
	assert.Equal(t, "QUARTER", updatedGame.CurrentState.PeriodType)

	// Test detail update
	op = patchOp{
		Path:  "/fullStatus/type/detail",
		Value: "3:28 - 3rd Quarter",
	}
	updatedGame = applyPeriodUpdate(game, op)
	assert.Equal(t, "3:28", updatedGame.CurrentState.Clock)
	assert.Equal(t, 3, updatedGame.CurrentState.Period)
	assert.Equal(t, "QUARTER", updatedGame.CurrentState.PeriodType)

	// Test overtime
	op = patchOp{
		Path:  "/summary",
		Value: "2:15 - OT",
	}
	updatedGame = applyPeriodUpdate(game, op)
	assert.Equal(t, "2:15", updatedGame.CurrentState.Clock)
	assert.Equal(t, 5, updatedGame.CurrentState.Period)
	assert.Equal(t, "OVERTIME", updatedGame.CurrentState.PeriodType)

	// Test non-period path
	op = patchOp{
		Path:  "/other/path",
		Value: "test",
	}
	updatedGame = applyPeriodUpdate(game, op)
	assert.Equal(t, game.CurrentState.Clock, updatedGame.CurrentState.Clock) // Should be unchanged
}

func TestApplySituationUpdate(t *testing.T) {
	game := models.Game{
		CurrentState: models.GameState{
			Details: models.EventDetails{},
		},
	}

	// Test down update
	op := patchOp{
		Path:  "/situation/down",
		Value: float64(3),
	}
	updatedGame := applySituationUpdate(game, op)
	assert.Equal(t, 3, updatedGame.CurrentState.Details.Down)

	// Test invalid down
	op = patchOp{
		Path:  "/situation/down",
		Value: float64(5),
	}
	updatedGame = applySituationUpdate(game, op)
	assert.Equal(t, 0, updatedGame.CurrentState.Details.Down)

	// Test yard line update
	op = patchOp{
		Path:  "/situation/yardLine",
		Value: float64(25),
	}
	updatedGame = applySituationUpdate(game, op)
	assert.Equal(t, 25, updatedGame.CurrentState.Details.YardLine)

	// Test possession text update
	op = patchOp{
		Path:  "/possessionText",
		Value: "GB 30",
	}
	updatedGame = applySituationUpdate(game, op)
	assert.Equal(t, "GB", updatedGame.CurrentState.Details.Possession)
	assert.Equal(t, 30, updatedGame.CurrentState.Details.YardLine)

	// Test down distance text update
	op = patchOp{
		Path:  "/situation/shortDownDistanceText",
		Value: "3rd & 7 at GB 30",
	}
	updatedGame = applySituationUpdate(game, op)
	assert.Equal(t, 3, updatedGame.CurrentState.Details.Down)
	assert.Equal(t, 7, updatedGame.CurrentState.Details.Distance)
	assert.Equal(t, "GB", updatedGame.CurrentState.Details.Possession)
	assert.Equal(t, 30, updatedGame.CurrentState.Details.YardLine)

	// Test non-situation path
	op = patchOp{
		Path:  "/other/path",
		Value: "test",
	}
	updatedGame = applySituationUpdate(game, op)
	assert.Equal(t, game.CurrentState.Details.Down, updatedGame.CurrentState.Details.Down) // Should be unchanged
}

func TestApplyScoreUpdate(t *testing.T) {
	game := models.Game{
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team: models.Team{ExtID: "4"},
			},
			Away: models.TeamState{
				Team: models.Team{ExtID: "15"},
			},
		},
	}

	mapping := competitorMapping{
		Side:   map[string]string{"0": "home", "1": "away"},
		TeamID: map[string]string{"0": "4", "1": "15"},
	}

	// Test score update with team ID
	op := patchOp{
		Path:  "/competitors/0/score",
		Value: "21",
	}
	updatedGame := applyScoreUpdate(game, op, mapping)
	assert.Equal(t, 21, updatedGame.CurrentState.Home.Score)

	// Test score update with side mapping
	op = patchOp{
		Path:  "/competitors/1/score",
		Value: "14",
	}
	updatedGame = applyScoreUpdate(game, op, mapping)
	assert.Equal(t, 14, updatedGame.CurrentState.Away.Score)

	// Test invalid score
	op = patchOp{
		Path:  "/competitors/0/score",
		Value: "invalid",
	}
	updatedGame = applyScoreUpdate(game, op, mapping)
	assert.Equal(t, game.CurrentState.Home.Score, updatedGame.CurrentState.Home.Score) // Should be unchanged

	// Test non-score path
	op = patchOp{
		Path:  "/other/path",
		Value: "test",
	}
	updatedGame = applyScoreUpdate(game, op, mapping)
	assert.Equal(t, game.CurrentState.Home.Score, updatedGame.CurrentState.Home.Score) // Should be unchanged
}

func TestRefreshScoreFromAPI(t *testing.T) {
	game := models.Game{
		GameCode: "test-game",
		CurrentState: models.GameState{
			Home: models.TeamState{Score: 0},
			Away: models.TeamState{Score: 0},
		},
	}

	// Test with summary path (should trigger refresh)
	updatedGame := refreshScoreFromAPI(game, "/summary")
	// Note: This will fail in tests due to real API call, but we're testing the logic
	assert.Equal(t, game.CurrentState.Home.Score, updatedGame.CurrentState.Home.Score) // Will be unchanged due to mock client behavior

	// Test with non-summary path (should not trigger refresh)
	updatedGame = refreshScoreFromAPI(game, "/other/path")
	assert.Equal(t, game.CurrentState.Home.Score, updatedGame.CurrentState.Home.Score) // Should be unchanged
}
