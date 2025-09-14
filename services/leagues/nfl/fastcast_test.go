package nfl

import (
	"encoding/json"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestFetchFastcastHost(t *testing.T) {
	// This test will fail in CI due to network access, but it's good to have locally
	// Skip if running in CI or if network is not available
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	host, err := fetchFastcastHost()
	if err != nil {
		t.Skipf("Skipping test due to network error: %v", err)
	}

	assert.NoError(t, err)
	assert.NotNil(t, host)
	assert.NotEmpty(t, host.IP)
	assert.Greater(t, host.SecurePort, 0)
	assert.NotEmpty(t, host.Token)
}

func TestStartNFLFastcast_Disabled(t *testing.T) {
	// Test when fastcast is disabled
	viper.Set("nfl.fastcast.enabled", false)
	defer viper.Set("nfl.fastcast.enabled", true)

	// This should return immediately without starting goroutine
	StartNFLFastcast()
}

func TestApplyNFLPatches_EmptyPayload(t *testing.T) {
	// Test with empty payload
	applyNFLPatches(json.RawMessage(`""`), "test-topic")
	// Should not panic
}

func TestApplyNFLPatches_InvalidJSON(t *testing.T) {
	// Test with invalid JSON
	applyNFLPatches(json.RawMessage(`invalid json`), "test-topic")
	// Should not panic
}

func TestApplyNFLPatches_Base64Zlib(t *testing.T) {
	// Test with base64+zlib encoded payload
	// This is a minimal valid base64+zlib encoded empty array
	payload := `"eJwLAAAAAA=="`
	applyNFLPatches(json.RawMessage(payload), "test-topic")
	// Should not panic
}

func TestApplyNFLPatches_DirectOps(t *testing.T) {
	// Test with direct ops array
	ops := `[{"op":"replace","path":"/test","value":"test"}]`
	applyNFLPatches(json.RawMessage(ops), "test-topic")
	// Should not panic
}

func TestApplyNFLPatches_WrapperWithOps(t *testing.T) {
	// Test with wrapper containing ops
	wrapper := `{"ts":1234567890,"~c":0,"pl":[{"op":"replace","path":"/test","value":"test"}]}`
	applyNFLPatches(json.RawMessage(wrapper), "test-topic")
	// Should not panic
}

func TestApplyNFLPatches_WrapperWithBase64(t *testing.T) {
	// Test with wrapper containing base64+zlib
	wrapper := `{"ts":1234567890,"~c":0,"pl":"eJwLAAAAAA=="}`
	applyNFLPatches(json.RawMessage(wrapper), "test-topic")
	// Should not panic
}

func TestApplyNFLPatches_InvalidBase64(t *testing.T) {
	// Test with invalid base64
	payload := `"invalid-base64"`
	applyNFLPatches(json.RawMessage(payload), "test-topic")
	// Should not panic
}

func TestApplyNFLPatches_InvalidZlib(t *testing.T) {
	// Test with valid base64 but invalid zlib
	payload := `"SGVsbG8gV29ybGQ="` // "Hello World" in base64
	applyNFLPatches(json.RawMessage(payload), "test-topic")
	// Should not panic
}

func TestApplyNFLPatches_EmptyOps(t *testing.T) {
	// Test with empty ops array
	ops := `[]`
	applyNFLPatches(json.RawMessage(ops), "test-topic")
	// Should not panic
}

func TestApplyNFLPatches_WithEventID(t *testing.T) {
	// Test with event ID in path
	ops := `[{"op":"replace","path":"/e:12345/clock","value":"15:00"}]`
	applyNFLPatches(json.RawMessage(ops), "test-topic")
	// Should not panic
}

func TestApplyNFLPatches_WithTopicEventID(t *testing.T) {
	// Test with event ID from topic
	ops := `[{"op":"replace","path":"/clock","value":"15:00"}]`
	applyNFLPatches(json.RawMessage(ops), "gp-football-nfl-12345")
	// Should not panic
}

func TestApplyNFLPatches_ClockUpdate(t *testing.T) {
	// Test clock update
	ops := `[{"op":"replace","path":"/e:12345/fullStatus/displayClock","value":"15:00"}]`
	applyNFLPatches(json.RawMessage(ops), "gp-football-nfl-12345")
	// Should not panic
}

func TestApplyNFLPatches_PeriodUpdate(t *testing.T) {
	// Test period update
	ops := `[{"op":"replace","path":"/e:12345/fullStatus/type/shortDetail","value":"15:00 - 3rd"}]`
	applyNFLPatches(json.RawMessage(ops), "gp-football-nfl-12345")
	// Should not panic
}

func TestApplyNFLPatches_ClockUpdate2(t *testing.T) {
	// Test clock update via displayClock path
	payload := `[{"op":"replace","path":"/fullStatus/displayClock","value":"12:34"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Verify clock was updated (this would require mocking memory store)
	// For now, just ensure no panic occurs
}

func TestApplyNFLPatches_PeriodUpdate2(t *testing.T) {
	// Test period update via shortDetail path
	payload := `[{"op":"replace","path":"/fullStatus/type/shortDetail","value":"6:14 - 3rd"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Verify period was updated (this would require mocking memory store)
	// For now, just ensure no panic occurs
}

func TestApplyNFLPatches_DownUpdate(t *testing.T) {
	// Test down update via situation/down path
	payload := `[{"op":"replace","path":"/situation/down","value":2}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Verify down was updated (this would require mocking memory store)
	// For now, just ensure no panic occurs
}

func TestApplyNFLPatches_YardLineUpdate(t *testing.T) {
	// Test yard line update via situation/yardLine path
	payload := `[{"op":"replace","path":"/situation/yardLine","value":25}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Verify yard line was updated (this would require mocking memory store)
	// For now, just ensure no panic occurs
}

func TestApplyNFLPatches_PossessionUpdate(t *testing.T) {
	// Test possession update via possessionText path
	payload := `[{"op":"replace","path":"/possessionText","value":"BUF 25"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Verify possession was updated (this would require mocking memory store)
	// For now, just ensure no panic occurs
}

func TestApplyNFLPatches_DownDistanceUpdate(t *testing.T) {
	// Test down and distance update via shortDownDistanceText path
	payload := `[{"op":"replace","path":"/situation/shortDownDistanceText","value":"2nd & 7 at BUF 25"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Verify down and distance were updated (this would require mocking memory store)
	// For now, just ensure no panic occurs
}

func TestApplyNFLPatches_SummaryUpdate(t *testing.T) {
	// Test summary update via summary path
	payload := `[{"op":"replace","path":"/summary","value":"3:28 - 3rd Quarter"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Verify summary was updated (this would require mocking memory store)
	// For now, just ensure no panic occurs
}

func TestApplyNFLPatches_LastPlayUpdate(t *testing.T) {
	// Test last play update via situation/lastPlay path
	payload := `[{"op":"replace","path":"/situation/lastPlay","value":"Touchdown pass"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Verify last play was updated (this would require mocking memory store)
	// For now, just ensure no panic occurs
}

func TestApplyNFLPatches_InvalidDown(t *testing.T) {
	// Test invalid down value (should set to 0)
	payload := `[{"op":"replace","path":"/situation/down","value":5}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Verify down was set to 0 (this would require mocking memory store)
	// For now, just ensure no panic occurs
}

func TestApplyNFLPatches_InvalidDownDistance(t *testing.T) {
	// Test invalid down in downDistanceText (should set to 0)
	payload := `[{"op":"replace","path":"/situation/shortDownDistanceText","value":"5th & 7 at BUF 25"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Verify down was set to 0 (this would require mocking memory store)
	// For now, just ensure no panic occurs
}

func TestApplyNFLPatches_EmptyValue(t *testing.T) {
	// Test empty value (should not update)
	payload := `[{"op":"replace","path":"/fullStatus/displayClock","value":""}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Verify no update occurred (this would require mocking memory store)
	// For now, just ensure no panic occurs
}

func TestApplyNFLPatches_InvalidBase642(t *testing.T) {
	// Test invalid base64 in wrapper
	payload := `{"ts":1234567890,"~c":1,"pl":"invalid-base64-string"}`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_InvalidZlib2(t *testing.T) {
	// Test invalid zlib data
	invalidZlib := "not-zlib-data"
	payload := `{"ts":1234567890,"~c":1,"pl":"` + invalidZlib + `"}`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_InvalidJSON2(t *testing.T) {
	// Test invalid JSON in wrapper
	payload := `{"ts":1234567890,"~c":1,"pl":"invalid-json"}`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_EmptyWrapper(t *testing.T) {
	// Test empty wrapper
	payload := `{"ts":1234567890,"~c":1,"pl":""}`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_NoWrapper(t *testing.T) {
	// Test direct JSON array without wrapper
	payload := `[{"op":"replace","path":"/test","value":"test"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_InvalidDirectJSON(t *testing.T) {
	// Test invalid direct JSON
	payload := `invalid-json`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_EmptyPayload2(t *testing.T) {
	// Test empty payload
	payload := ``

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_InvalidTopic2(t *testing.T) {
	// Test invalid topic format
	payload := `[{"op":"replace","path":"/test","value":"test"}]`

	applyNFLPatches(json.RawMessage(payload), "invalid-topic-format")

	// Should not panic
}

func TestApplyNFLPatches_NoEventID(t *testing.T) {
	// Test payload with no event ID in path or topic
	payload := `[{"op":"replace","path":"/test","value":"test"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-")

	// Should not panic
}

func TestApplyNFLPatches_DirectBase64Zlib(t *testing.T) {
	// Test direct base64+zlib path
	// Use a simple base64 encoded string for testing
	encoded := "eyJvcCI6InJlcGxhY2UiLCJwYXRoIjoiL3Rlc3QiLCJ2YWx1ZSI6InRlc3QifQ=="

	applyNFLPatches(json.RawMessage(encoded), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_CompetitorMapping(t *testing.T) {
	// Test competitor mapping logic
	payload := `[{"op":"replace","path":"/competitors/0/homeAway","value":"home"},{"op":"replace","path":"/competitors/0/team/id","value":"15"},{"op":"replace","path":"/competitors/0/score","value":"21"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_EventIDExtraction(t *testing.T) {
	// Test event ID extraction from path
	payload := `[{"op":"replace","path":"/e:12345/test","value":"test"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_InvalidCompetitorData(t *testing.T) {
	// Test with invalid competitor data
	payload := `[{"op":"replace","path":"/competitors/0/homeAway","value":123}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_InvalidTeamID(t *testing.T) {
	// Test with invalid team ID
	payload := `[{"op":"replace","path":"/competitors/0/team/id","value":123}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_ClockSeconds(t *testing.T) {
	// Test clock update via fullStatus/clock path (seconds to M:SS conversion)
	payload := `[{"op":"replace","path":"/fullStatus/clock","value":900}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_PeriodTypeQuarter(t *testing.T) {
	// Test period type setting for quarters
	payload := `[{"op":"replace","path":"/fullStatus/type/shortDetail","value":"6:14 - 2nd"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_PeriodTypeOvertime(t *testing.T) {
	// Test period type setting for overtime
	payload := `[{"op":"replace","path":"/fullStatus/type/shortDetail","value":"6:14 - OT"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_PeriodType2OT(t *testing.T) {
	// Test period type setting for 2nd overtime
	payload := `[{"op":"replace","path":"/fullStatus/type/shortDetail","value":"6:14 - 2OT"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_DetailPath(t *testing.T) {
	// Test period update via fullStatus/type/detail path
	payload := `[{"op":"replace","path":"/fullStatus/type/detail","value":"3:28 - 3rd Quarter"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_YardLineFloat(t *testing.T) {
	// Test yard line update with float value
	payload := `[{"op":"replace","path":"/situation/yardLine","value":25.0}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_PossessionTextWithYardLine(t *testing.T) {
	// Test possession text with yard line parsing
	payload := `[{"op":"replace","path":"/possessionText","value":"BUF 25"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_DownDistanceText(t *testing.T) {
	// Test down and distance text parsing
	payload := `[{"op":"replace","path":"/downDistanceText","value":"2nd & 7 at BUF 25"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_ScoreUpdate(t *testing.T) {
	// Test score update via competitors path
	payload := `[{"op":"replace","path":"/competitors/0/score","value":"21"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_LastPlayPath(t *testing.T) {
	// Test last play path
	payload := `[{"op":"replace","path":"/situation/lastPlay","value":"Touchdown pass"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_ScoreUpdateWithTeamID(t *testing.T) {
	// Test score update with team ID mapping
	payload := `[{"op":"replace","path":"/competitors/0/team/id","value":"4"},{"op":"replace","path":"/competitors/0/score","value":"21"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_ScoreUpdateWithSide(t *testing.T) {
	// Test score update with home/away side mapping
	payload := `[{"op":"replace","path":"/competitors/0/homeAway","value":"home"},{"op":"replace","path":"/competitors/0/score","value":"21"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_ScoreUpdateUnknownSide(t *testing.T) {
	// Test score update with unknown side (should trigger fallback)
	payload := `[{"op":"replace","path":"/competitors/0/score","value":"21"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_SummaryRefresh(t *testing.T) {
	// Test summary refresh logic
	payload := `[{"op":"replace","path":"/summary","value":"Touchdown!"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_LastPlayRefresh(t *testing.T) {
	// Test last play refresh logic
	payload := `[{"op":"replace","path":"/situation/lastPlay","value":"Touchdown pass"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_InvalidScoreValue(t *testing.T) {
	// Test with invalid score value (non-numeric)
	payload := `[{"op":"replace","path":"/competitors/0/score","value":"invalid"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_EmptyScoreValue(t *testing.T) {
	// Test with empty score value
	payload := `[{"op":"replace","path":"/competitors/0/score","value":""}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_CompetitorIndexMapping(t *testing.T) {
	// Test competitor index mapping for multiple competitors
	payload := `[{"op":"replace","path":"/competitors/0/homeAway","value":"home"},{"op":"replace","path":"/competitors/1/homeAway","value":"away"},{"op":"replace","path":"/competitors/0/score","value":"21"},{"op":"replace","path":"/competitors/1/score","value":"14"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_InvalidCompetitorIndex(t *testing.T) {
	// Test with invalid competitor index
	payload := `[{"op":"replace","path":"/competitors/99/homeAway","value":"home"}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_NonStringValue(t *testing.T) {
	// Test with non-string value for homeAway
	payload := `[{"op":"replace","path":"/competitors/0/homeAway","value":123}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_NonStringTeamID(t *testing.T) {
	// Test with non-string value for team ID
	payload := `[{"op":"replace","path":"/competitors/0/team/id","value":123}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_EmptyHomeAwayValue(t *testing.T) {
	// Test with empty homeAway value
	payload := `[{"op":"replace","path":"/competitors/0/homeAway","value":""}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}

func TestApplyNFLPatches_EmptyTeamIDValue(t *testing.T) {
	// Test with empty team ID value
	payload := `[{"op":"replace","path":"/competitors/0/team/id","value":""}]`

	applyNFLPatches(json.RawMessage(payload), "gp-football-nfl-test-game")

	// Should not panic
}
