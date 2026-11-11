package homeassistant

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"goalfeed/models"

	"github.com/spf13/viper"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type haState struct {
	State      string                 `json:"state"`
	Attributes map[string]interface{} `json:"attributes"`
}

// recordingHA captures every /api/states/<entity_id> POST, keyed by entity id.
func recordingHA(t *testing.T) (map[string]haState, *sync.Mutex) {
	t.Helper()
	got := map[string]haState{}
	mu := &sync.Mutex{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id, ok := strings.CutPrefix(r.URL.Path, "/core/api/states/"); ok {
			b, _ := io.ReadAll(r.Body)
			var s haState
			_ = json.Unmarshal(b, &s)
			mu.Lock()
			got[id] = s
			mu.Unlock()
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	os.Setenv("SUPERVISOR_API", srv.URL)
	os.Setenv("SUPERVISOR_TOKEN", "t")
	t.Cleanup(func() {
		os.Unsetenv("SUPERVISOR_API")
		os.Unsetenv("SUPERVISOR_TOKEN")
	})
	debounceAfter = 0
	entityCache = map[string]entityCacheEntry{}
	return got, mu
}

func liveNHLGame() models.Game {
	return models.Game{
		GameCode: "2026020291",
		LeagueId: models.LeagueIdNHL,
		CurrentState: models.GameState{
			Status:     models.StatusActive,
			Period:     3,
			PeriodType: "REGULAR",
			Clock:      "01:46",
			Venue:      models.Venue{Name: "Canada Life Centre"},
			Home: models.TeamState{
				Team: models.Team{TeamCode: "WPG"}, Score: 3,
				PeriodScores: []int{2, 0, 1}, PeriodShots: []int{8, 5, 10},
				Statistics: models.TeamStats{
					Shots: 23, HasTeamGameStats: true, Hits: 22, Faceoffs: 27, FaceoffTotal: 51,
					FaceoffPct: 52.9, PowerPlayGoals: 2, PowerPlays: 4, PowerPlayPct: 50,
					PenaltyMinutes: 8, BlockedShots: 11, Giveaways: 9, Takeaways: 7,
				},
			},
			Away: models.TeamState{
				Team: models.Team{TeamCode: "NJD"}, Score: 2,
				Statistics: models.TeamStats{Shots: 21, HasTeamGameStats: true, Hits: 8},
			},
			Situation: &models.Situation{Code: "1451", AwaySkaters: 4, HomeSkaters: 5, PowerPlayTeam: "WPG", TimeRemaining: "01:12"},
			LastGoal: &models.GoalSummary{
				TeamCode: "WPG", Scorer: "Kyle Connor", Assists: []string{"Mark Scheifele"},
				Period: 3, TimeInPeriod: "17:01", Strength: "PP", HomeScore: 3, AwayScore: 2,
			},
		},
	}
}

func TestPublishTeamSensors_NHLTeamGameStats(t *testing.T) {
	got, _ := recordingHA(t)
	PublishTeamSensors(liveNHLGame())

	for entity, want := range map[string]string{
		"sensor.goalfeed_nhl_wpg_team_hits":          "22",
		"sensor.goalfeed_nhl_wpg_team_faceoff_wins":  "27",
		"sensor.goalfeed_nhl_wpg_team_faceoff_pct":   "52.9",
		"sensor.goalfeed_nhl_wpg_team_pim":           "8",
		"sensor.goalfeed_nhl_wpg_team_blocked_shots": "11",
		"sensor.goalfeed_nhl_wpg_team_giveaways":     "9",
		"sensor.goalfeed_nhl_wpg_team_takeaways":     "7",
		"sensor.goalfeed_nhl_wpg_team_power_play":    "2/4",
		"sensor.goalfeed_nhl_njd_team_hits":          "8",
	} {
		require.Contains(t, got, entity)
		assert.Equal(t, want, got[entity].State, entity)
	}
	pp := got["sensor.goalfeed_nhl_wpg_team_power_play"].Attributes
	assert.Equal(t, 2.0, pp["goals"])
	assert.Equal(t, 4.0, pp["opportunities"])
	assert.Equal(t, 50.0, pp["pct"])
	assert.Equal(t, 51.0, got["sensor.goalfeed_nhl_wpg_team_faceoff_wins"].Attributes["faceoffs_taken"])
}

func TestPublishTeamSensors_NHLSituationBinarySensors(t *testing.T) {
	got, _ := recordingHA(t)
	PublishTeamSensors(liveNHLGame())

	wpgPP := got["binary_sensor.goalfeed_nhl_wpg_team_on_power_play"]
	assert.Equal(t, "on", wpgPP.State)
	assert.Equal(t, "5v4", wpgPP.Attributes["strength"])
	assert.Equal(t, "01:12", wpgPP.Attributes["time_remaining"])

	njdPP := got["binary_sensor.goalfeed_nhl_njd_team_on_power_play"]
	assert.Equal(t, "off", njdPP.State)
	assert.Equal(t, "4v5", njdPP.Attributes["strength"])

	assert.Equal(t, "off", got["binary_sensor.goalfeed_nhl_wpg_team_empty_net"].State)
	assert.Equal(t, "off", got["binary_sensor.goalfeed_nhl_wpg_team_intermission"].State)
}

func TestPublishTeamSensors_NHLEmptyNetIsPerTeam(t *testing.T) {
	got, _ := recordingHA(t)
	g := liveNHLGame()
	g.CurrentState.Situation = &models.Situation{Code: "1560", AwaySkaters: 5, HomeSkaters: 6, HomeEmptyNet: true}
	g.CurrentState.InIntermission = false
	PublishTeamSensors(g)

	assert.Equal(t, "on", got["binary_sensor.goalfeed_nhl_wpg_team_empty_net"].State, "WPG pulled its goalie")
	assert.Equal(t, "off", got["binary_sensor.goalfeed_nhl_njd_team_empty_net"].State)
	assert.Equal(t, "off", got["binary_sensor.goalfeed_nhl_wpg_team_on_power_play"].State, "an extra attacker is not a power play")
}

func TestPublishTeamSensors_NHLGameSensor(t *testing.T) {
	got, _ := recordingHA(t)
	PublishTeamSensors(liveNHLGame())

	gs, ok := got["sensor.goalfeed_nhl_wpg_game"]
	require.True(t, ok, "one game sensor per team")
	assert.Equal(t, "active", gs.State)
	a := gs.Attributes
	assert.Equal(t, "NJD", a["opponent"])
	assert.Equal(t, "home", a["home_away"])
	assert.Equal(t, 3.0, a["score"])
	assert.Equal(t, 2.0, a["opponent_score"])
	assert.NotContains(t, a, "clock", "a per-second attribute would write a recorder row every second")
	assert.Equal(t, []interface{}{2.0, 0.0, 1.0}, a["goals_by_period"])
	assert.Equal(t, []interface{}{8.0, 5.0, 10.0}, a["shots_by_period"])
	assert.Equal(t, []interface{}{}, a["opponent_goals_by_period"])

	sit := a["situation"].(map[string]interface{})
	assert.Equal(t, "5v4", sit["strength"])
	assert.Equal(t, "WPG", sit["power_play_team"])

	lg := a["last_goal"].(map[string]interface{})
	assert.Equal(t, "Kyle Connor", lg["scorer"])
	assert.Equal(t, []interface{}{"Mark Scheifele"}, lg["assists"])
	assert.Equal(t, "PP", lg["strength"])

	njd := got["sensor.goalfeed_nhl_njd_game"].Attributes
	assert.Equal(t, "WPG", njd["opponent"])
	assert.Equal(t, "away", njd["home_away"])
	assert.Equal(t, "4v5", njd["situation"].(map[string]interface{})["strength"])
}

func TestPublishTeamSensors_NHLStatsHeldUntilPanelRead(t *testing.T) {
	got, _ := recordingHA(t)
	g := liveNHLGame()
	g.CurrentState.Home.Statistics = models.TeamStats{Shots: 4}
	PublishTeamSensors(g)

	assert.NotContains(t, got, "sensor.goalfeed_nhl_wpg_team_hits", "no right-rail yet: leave the previous values alone")
	assert.Equal(t, "4", got["sensor.goalfeed_nhl_wpg_team_shots"].State)
}

func TestPublishTeamSensors_NHLUnknownSituationIsOff(t *testing.T) {
	got, _ := recordingHA(t)
	g := liveNHLGame()
	g.CurrentState.Situation = nil
	g.CurrentState.InIntermission = true
	PublishTeamSensors(g)

	assert.Equal(t, "off", got["binary_sensor.goalfeed_nhl_wpg_team_on_power_play"].State)
	assert.Equal(t, "off", got["binary_sensor.goalfeed_nhl_wpg_team_empty_net"].State)
	assert.Equal(t, "on", got["binary_sensor.goalfeed_nhl_wpg_team_intermission"].State)
	assert.Nil(t, got["sensor.goalfeed_nhl_wpg_game"].Attributes["situation"])
}

func TestPublishEndOfGameReset_NHLSituationOff(t *testing.T) {
	got, _ := recordingHA(t)
	g := liveNHLGame()
	PublishTeamSensors(g)
	require.Equal(t, "on", got["binary_sensor.goalfeed_nhl_wpg_team_on_power_play"].State)

	g.CurrentState.Status = models.StatusEnded
	PublishEndOfGameReset(g)
	assert.Equal(t, "off", got["binary_sensor.goalfeed_nhl_wpg_team_on_power_play"].State)
	assert.Equal(t, "22", got["sensor.goalfeed_nhl_wpg_team_hits"].State, "final stats are kept")
}

func TestPublishTeamSensors_NonNHLHockeyGetsNoNewEntities(t *testing.T) {
	got, _ := recordingHA(t)
	g := liveNHLGame()
	g.LeagueId = models.LeagueIdOlympicMensHockey
	PublishTeamSensors(g)
	assert.NotContains(t, got, "sensor.goalfeed_olympic_men_wpg_game")
	assert.NotContains(t, got, "binary_sensor.goalfeed_olympic_men_wpg_team_on_power_play")
	assert.NotContains(t, got, "sensor.goalfeed_olympic_men_wpg_team_hits")
	assert.Contains(t, got, "sensor.goalfeed_olympic_men_wpg_team_shots", "existing hockey sensors unchanged")
}

func TestPublishBaseline_NHLNewEntities(t *testing.T) {
	got, _ := recordingHA(t)
	viper.Set("watch.nhl", []string{"WPG"})
	t.Cleanup(func() { viper.Set("watch.nhl", nil) })
	PublishBaselineForMonitoredTeams()

	assert.Equal(t, "idle", got["sensor.goalfeed_nhl_wpg_game"].State)
	assert.Equal(t, "off", got["binary_sensor.goalfeed_nhl_wpg_team_on_power_play"].State)
	assert.Equal(t, "off", got["binary_sensor.goalfeed_nhl_wpg_team_empty_net"].State)
	assert.Equal(t, "off", got["binary_sensor.goalfeed_nhl_wpg_team_intermission"].State)
}
