package nhl

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// right_rail_2026020277_final.json is a recorded response from
// /v1/gamecenter/2026020277/right-rail (FLA @ BOS, final, 2026-11-09),
// with only the logo URL keys removed (AGENTS.md: no logos in the repo).
func loadRightRailFixture(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/right_rail_2026020277_final.json")
	require.NoError(t, err)
	return b
}

func TestGetNHLRightRail_RecordedFixture(t *testing.T) {
	payload := loadRightRailFixture(t)
	old := fetchByte
	var gotURL string
	fetchByte = func(url string, ret chan []byte) { gotURL = url; ret <- payload }
	defer func() { fetchByte = old }()

	resp, err := NHLApiClient{}.GetNHLRightRail("2026020277")
	require.NoError(t, err)
	assert.Equal(t, "https://api-web.nhle.com/v1/gamecenter/2026020277/right-rail", gotURL)

	away, home := resp.TeamStats()
	assert.True(t, away.Present)
	assert.Empty(t, away.UnknownCategories)

	assert.Equal(t, NHLTeamGameStats{
		Present: true, ShotsOnGoal: 42, FaceoffWins: 28, FaceoffTotal: 77, FaceoffPct: 36.4,
		PowerPlayGoals: 1, PowerPlayOpps: 5, PowerPlayPct: 20, PenaltyMinutes: 72,
		Hits: 38, BlockedShots: 6, Giveaways: 13, Takeaways: 5,
	}, away)
	assert.Equal(t, NHLTeamGameStats{
		Present: true, ShotsOnGoal: 37, FaceoffWins: 49, FaceoffTotal: 77, FaceoffPct: 63.6,
		PowerPlayGoals: 1, PowerPlayOpps: 9, PowerPlayPct: 11.1, PenaltyMinutes: 54,
		Hits: 13, BlockedShots: 6, Giveaways: 20, Takeaways: 2,
	}, home)

	require.Len(t, resp.ShotsByPeriod, 3)
	assert.Equal(t, FlexInt(19), resp.ShotsByPeriod[0].Away)
	assert.Equal(t, FlexInt(18), resp.ShotsByPeriod[1].Home)
	require.Len(t, resp.Linescore.ByPeriod, 3)
	assert.Equal(t, FlexInt(3), resp.Linescore.ByPeriod[0].Home)
	assert.Equal(t, FlexInt(4), resp.Linescore.Totals.Home)
	assert.Equal(t, FlexInt(1), resp.Linescore.Totals.Away)
}

// The upstream mixes value types per category. Swapping every value to the
// other encoding (numbers as strings, and so on) must not lose the panel.
func TestRightRailTeamStats_ToleratesTypeDrift(t *testing.T) {
	raw := `{"teamGameStats":[
		{"category":"sog","awayValue":"12","homeValue":9.0},
		{"category":"faceoffWinningPctg","awayValue":"0.5","homeValue":55.5},
		{"category":"powerPlay","awayValue":" 2 / 4 ","homeValue":3},
		{"category":"hits","awayValue":null,"homeValue":{"oops":1}},
		{"category":"xGoals","awayValue":1.2,"homeValue":0.8}
	],"shotsByPeriod":[{"away":"7","home":5}]}`
	var resp NHLRightRailResponse
	require.NoError(t, json.Unmarshal([]byte(raw), &resp))

	away, home := resp.TeamStats()
	assert.Equal(t, 12, away.ShotsOnGoal)
	assert.Equal(t, 9, home.ShotsOnGoal)
	assert.Equal(t, 50.0, away.FaceoffPct)
	assert.Equal(t, 55.5, home.FaceoffPct, "a value already in 0-100 is kept")
	assert.Equal(t, 2, away.PowerPlayGoals)
	assert.Equal(t, 4, away.PowerPlayOpps)
	assert.Equal(t, 0, home.PowerPlayOpps, "a non-ratio power play value reads as zero, not a crash")
	assert.Equal(t, 0, away.Hits)
	assert.Equal(t, []string{"xGoals"}, away.UnknownCategories)
	assert.Equal(t, FlexInt(7), resp.ShotsByPeriod[0].Away)
}

func TestRightRailTeamStats_EmptyPanelIsNotPresent(t *testing.T) {
	away, home := NHLRightRailResponse{}.TeamStats()
	assert.False(t, away.Present)
	assert.False(t, home.Present)
}

func TestGetNHLRightRail_EmptyBodyIsAnError(t *testing.T) {
	defer withStubFetchNHLError(t)()
	_, err := NHLApiClient{}.GetNHLRightRail("1")
	assert.Error(t, err)
}

func TestGetNHLRightRail_GarbageBodyIsAnError(t *testing.T) {
	old := fetchByte
	fetchByte = func(url string, ret chan []byte) { ret <- []byte("<html>502</html>") }
	defer func() { fetchByte = old }()
	_, err := NHLApiClient{}.GetNHLRightRail("1")
	assert.Error(t, err)
}
