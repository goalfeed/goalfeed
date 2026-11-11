package nhl

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	nhlClients "goalfeed/clients/leagues/nhl"
	"goalfeed/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fixtureDir = "../../../clients/leagues/nhl/testdata/"

// fixtureClient serves recorded NHL responses.
type fixtureClient struct {
	nhlClients.MockNHLApiClient
	landing      nhlClients.NHLScoreboardResponse
	rightRail    nhlClients.NHLRightRailResponse
	rightRailErr error
	rightRailHit int
}

func (f *fixtureClient) GetNHLScoreBoard(string) nhlClients.NHLScoreboardResponse { return f.landing }
func (f *fixtureClient) GetNHLRightRail(string) (nhlClients.NHLRightRailResponse, error) {
	f.rightRailHit++
	return f.rightRail, f.rightRailErr
}

func loadJSON(t *testing.T, name string, v interface{}) {
	t.Helper()
	b, err := os.ReadFile(fixtureDir + name)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(b, v))
}

func TestSituationFromCode(t *testing.T) {
	cases := []struct {
		code         string
		pp           string
		home, away   string // strength from each side
		homeEN, awEN bool
	}{
		{"1551", "", "5v5", "5v5", false, false},
		{"1451", "WPG", "5v4", "4v5", false, false}, // home PP (recorded: 2026020291)
		{"1351", "WPG", "5v3", "3v5", false, false}, // home 5-on-3 (recorded: 2026020291)
		{"1541", "NJD", "4v5", "5v4", false, false}, // away PP
		{"1560", "", "6v5", "5v6", true, false},     // home goalie pulled: extra attacker, not a PP
		{"0651", "", "5v6", "6v5", false, true},     // away goalie pulled
		{"0551", "WPG", "5v5", "5v5", false, true},  // away pulled while shorthanded: still home PP
		{"1431", "NJD", "3v4", "4v3", false, false}, // 4-on-3 overtime PP (recorded: 2026020287)
		{"1331", "", "3v3", "3v3", false, false},    // overtime
	}
	for _, c := range cases {
		t.Run(c.code, func(t *testing.T) {
			s, ok := situationFromCode(c.code, "WPG", "NJD")
			require.True(t, ok)
			assert.Equal(t, c.pp, s.PowerPlayTeam)
			assert.Equal(t, c.home, s.Strength(true))
			assert.Equal(t, c.away, s.Strength(false))
			assert.Equal(t, c.homeEN, s.HomeEmptyNet)
			assert.Equal(t, c.awEN, s.AwayEmptyNet)
		})
	}
	for _, bad := range []string{"", "155", "15511", "15a1", "2551"} {
		_, ok := situationFromCode(bad, "WPG", "NJD")
		assert.False(t, ok, bad)
	}
}

func TestLastGoalFromRecordedLanding(t *testing.T) {
	var landing nhlClients.NHLScoreboardResponse
	loadJSON(t, "landing_2026020291_final.json", &landing)

	var st models.GameState
	applyLandingDetail(&st, landing)
	require.Len(t, st.Goals, 6)
	assert.Equal(t, "PP", st.Goals[1].Strength)
	assert.Equal(t, "WPG", st.Goals[1].TeamCode)
	g := st.LastGoal
	require.NotNil(t, g)
	assert.Equal(t, st.Goals[5], *g)
	assert.Equal(t, models.GoalSummary{
		EventId: "983", TeamCode: "WPG", Scorer: "Kyle Connor", ScorerId: "8478398",
		ScorerGoals: 15, Assists: []string{}, Period: 3, PeriodType: "REGULAR",
		TimeInPeriod: "19:02", Strength: "EV", EmptyNet: true, ShotType: "wrist",
		HomeScore: 4, AwayScore: 2,
	}, *g)
}

func TestGoalFor_MatchesEachPointToItsGoal(t *testing.T) {
	var landing nhlClients.NHLScoreboardResponse
	loadJSON(t, "landing_2026020291_final.json", &landing)
	var st models.GameState
	st.Home.Team.TeamCode, st.Away.Team.TeamCode = "WPG", "NJD"
	applyLandingDetail(&st, landing)

	// WPG (home) scored 1st, 2nd (PP), 3rd (PP) and 4th (empty net).
	g, ok := st.GoalFor("WPG", 4)
	require.True(t, ok)
	assert.True(t, g.EmptyNet)
	g, ok = st.GoalFor("WPG", 2)
	require.True(t, ok)
	assert.Equal(t, "PP", g.Strength)
	g, ok = st.GoalFor("NJD", 2)
	require.True(t, ok)
	assert.Equal(t, "NJD", g.TeamCode)

	_, ok = st.GoalFor("WPG", 5)
	assert.False(t, ok, "a goal the summary has not caught up with yet")
}

func TestApplyLandingDetail_FinalGameHasNoSituation(t *testing.T) {
	var landing nhlClients.NHLScoreboardResponse
	loadJSON(t, "landing_2026020291_final.json", &landing)
	var st models.GameState
	applyLandingDetail(&st, landing)
	assert.Nil(t, st.Situation)
	assert.False(t, st.InIntermission)
	assert.NotNil(t, st.LastGoal)
}

func TestApplyLandingDetail_LiveSituation(t *testing.T) {
	var landing nhlClients.NHLScoreboardResponse
	loadJSON(t, "landing_2026020291_final.json", &landing)

	landing.GameState = "LIVE"
	landing.Situation = nil
	var st models.GameState
	applyLandingDetail(&st, landing)
	require.NotNil(t, st.Situation, "a live game with no situation block is even strength")
	assert.Equal(t, "", st.Situation.PowerPlayTeam)
	assert.Equal(t, "5v5", st.Situation.Strength(true))

	landing.GameState = "CRIT"
	landing.Situation = &nhlClients.LandingSituation{SituationCode: "1451", TimeRemaining: "01:46"}
	applyLandingDetail(&st, landing)
	require.NotNil(t, st.Situation)
	assert.Equal(t, "WPG", st.Situation.PowerPlayTeam)
	assert.Equal(t, "01:46", st.Situation.TimeRemaining)

	landing.Clock.InIntermission = true
	applyLandingDetail(&st, landing)
	assert.True(t, st.InIntermission)
}

func TestGameUpdate_FillsRightRailStats(t *testing.T) {
	c := &fixtureClient{}
	loadJSON(t, "landing_2026020291_final.json", &c.landing)
	loadJSON(t, "right_rail_2026020291_final.json", &c.rightRail)
	svc := NHLService{Client: c}

	ch := make(chan models.GameUpdate, 1)
	svc.GetGameUpdate(models.Game{GameCode: "2026020291", LeagueId: models.LeagueIdNHL}, ch)
	u := <-ch

	// NJD @ WPG, 2026-11-10, final 2-4.
	assert.Equal(t, 1, c.rightRailHit)
	home := u.NewState.Home.Statistics
	assert.True(t, home.HasTeamGameStats)
	assert.Equal(t, 25, home.Shots)
	assert.Equal(t, 22, home.Hits)
	assert.Equal(t, 27, home.Faceoffs)
	assert.Equal(t, 51, home.FaceoffTotal)
	assert.Equal(t, 52.9, home.FaceoffPct)
	assert.Equal(t, 2, home.PowerPlayGoals)
	assert.Equal(t, 4, home.PowerPlays)
	assert.Equal(t, 50.0, home.PowerPlayPct)
	assert.Equal(t, 8, home.PenaltyMinutes)
	assert.Equal(t, 11, home.BlockedShots)
	assert.Equal(t, 9, home.Giveaways)
	assert.Equal(t, 7, home.Takeaways)
	assert.Equal(t, []int{2, 0, 2}, u.NewState.Home.PeriodScores)
	assert.Equal(t, []int{0, 2, 0}, u.NewState.Away.PeriodScores)
	assert.Equal(t, []int{11, 6, 4}, u.NewState.Away.PeriodShots)
	assert.Equal(t, 0, u.NewState.Away.Statistics.PowerPlayGoals)
	assert.Equal(t, 0.0, u.NewState.Away.Statistics.PowerPlayPct)
	assert.NotNil(t, u.NewState.LastGoal)
}

func TestGameUpdate_RightRailFailureKeepsPreviousStats(t *testing.T) {
	c := &fixtureClient{rightRailErr: errors.New("boom")}
	loadJSON(t, "landing_2026020291_final.json", &c.landing)
	svc := NHLService{Client: c}

	prev := models.Game{GameCode: "2026020291", LeagueId: models.LeagueIdNHL}
	prev.CurrentState.Home.Statistics = models.TeamStats{HasTeamGameStats: true, Hits: 11, Shots: 3}
	prev.CurrentState.Home.PeriodScores = []int{1}

	ch := make(chan models.GameUpdate, 1)
	svc.GetGameUpdate(prev, ch)
	u := <-ch

	assert.Equal(t, 11, u.NewState.Home.Statistics.Hits, "a failed right-rail must not zero the stats")
	assert.Equal(t, c.landing.HomeTeam.Sog, u.NewState.Home.Statistics.Shots, "shots still come from landing")
	assert.Equal(t, []int{1}, u.NewState.Home.PeriodScores)
}
