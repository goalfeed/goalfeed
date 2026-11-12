package nhl

import (
	"fmt"
	"goalfeed/clients/leagues/nhl"
	"goalfeed/models"
	"strconv"
	"strings"
	"sync"
	"time"
)

// rightRailInterval spaces out right-rail fetches for one game. Live games
// are polled every second for goal detection; the team stats panel does not
// need that, and an unofficial API hit at 1 Hz per game is how an IP gets
// blocked, which would take goal detection (landing) down with it. Situation,
// power play, empty net and goals come from landing and stay per-second.
var rightRailInterval = 15 * time.Second

var (
	rightRailMu        sync.Mutex
	rightRailLast      = map[string]time.Time{}
	rightRailFinalDone = map[string]bool{}
)

func isFinalGameState(state string) bool {
	return state == "FINAL" || state == "OFF"
}

// rightRailDue reports whether game gameCode should fetch right-rail now, and
// records the attempt if so. Every state is throttled to rightRailInterval,
// including ones goalfeed does not recognise, since those are still polled
// every second. The one exception is the first poll that sees the game
// final, which always fetches once so the final numbers are complete; later
// final polls never do.
func rightRailDue(gameCode string, gameState string, now time.Time) bool {
	rightRailMu.Lock()
	defer rightRailMu.Unlock()
	if isFinalGameState(gameState) {
		if rightRailFinalDone[gameCode] {
			return false
		}
		rightRailFinalDone[gameCode] = true
		delete(rightRailLast, gameCode)
		return true
	}
	if last, ok := rightRailLast[gameCode]; ok && now.Sub(last) < rightRailInterval {
		return false
	}
	rightRailLast[gameCode] = now
	return true
}

// isLiveGameState reports the landing gameState values during which an
// on-ice situation is meaningful. CRIT is the NHL's "critical" state for the
// late stages of a close game.
func isLiveGameState(state string) bool {
	return state == "LIVE" || state == "CRIT"
}

// applyLandingDetail fills the hockey situation, intermission flag and last
// goal from the landing response goalfeed already fetches every poll. It
// costs no extra request.
func applyLandingDetail(state *models.GameState, sb nhl.NHLScoreboardResponse) {
	state.InIntermission = sb.Clock.InIntermission
	state.Goals = goalsFromLanding(sb)
	state.LastGoal = nil
	if n := len(state.Goals); n > 0 {
		last := state.Goals[n-1]
		state.LastGoal = &last
	}

	if !isLiveGameState(sb.GameState) {
		state.Situation = nil
		return
	}
	code := "1551"
	timeRemaining := ""
	if sb.Situation != nil && sb.Situation.SituationCode != "" {
		code = sb.Situation.SituationCode
		timeRemaining = sb.Situation.TimeRemaining
	}
	sit, ok := situationFromCode(code, sb.HomeTeam.Abbrev, sb.AwayTeam.Abbrev)
	if !ok {
		logger.Warn(fmt.Sprintf("NHL game %d: unrecognised situationCode %q; treating as unknown", sb.ID, code))
		state.Situation = nil
		return
	}
	sit.TimeRemaining = timeRemaining
	state.Situation = &sit
}

// situationFromCode decodes an NHL situationCode. The four digits are, in
// order: away goalie in net (1/0), away skaters, home skaters, home goalie in
// net. "1451" is a home power play at 5-on-4; "1560" is 6-on-5 with the home
// goalie pulled.
//
// A pulled goalie's extra attacker is not a man advantage, so the power play
// team is decided on skaters net of any extra attacker: "1560" is even
// strength, and "0551" (away goalie pulled while shorthanded, so 5 skaters a
// side) is still a home power play.
func situationFromCode(code, homeCode, awayCode string) (models.Situation, bool) {
	if len(code) != 4 {
		return models.Situation{}, false
	}
	digits := [4]int{}
	for i, r := range code {
		if r < '0' || r > '9' {
			return models.Situation{}, false
		}
		digits[i] = int(r - '0')
	}
	awayGoalie, awaySkaters, homeSkaters, homeGoalie := digits[0], digits[1], digits[2], digits[3]
	if awayGoalie > 1 || homeGoalie > 1 {
		return models.Situation{}, false
	}
	s := models.Situation{
		Code:         code,
		AwaySkaters:  awaySkaters,
		HomeSkaters:  homeSkaters,
		AwayEmptyNet: awayGoalie == 0,
		HomeEmptyNet: homeGoalie == 0,
	}
	awayNet, homeNet := awaySkaters, homeSkaters
	if s.AwayEmptyNet {
		awayNet--
	}
	if s.HomeEmptyNet {
		homeNet--
	}
	switch {
	case homeNet > awayNet:
		s.PowerPlayTeam = homeCode
	case awayNet > homeNet:
		s.PowerPlayTeam = awayCode
	}
	return s, true
}

func goalsFromLanding(sb nhl.NHLScoreboardResponse) []models.GoalSummary {
	var goals []models.GoalSummary
	for _, period := range sb.Summary.Scoring {
		for _, g := range period.Goals {
			goals = append(goals, goalSummary(sb, period.PeriodDescriptor, g))
		}
	}
	return goals
}

func goalSummary(sb nhl.NHLScoreboardResponse, pd nhl.PeriodDescriptor, g nhl.LandingGoal) models.GoalSummary {
	assists := make([]string, 0, len(g.Assists))
	for _, a := range g.Assists {
		assists = append(assists, playerName(a.Name, a.FirstName, a.LastName))
	}
	team := g.TeamAbbrev.Default
	if team == "" {
		if g.IsHome {
			team = sb.HomeTeam.Abbrev
		} else {
			team = sb.AwayTeam.Abbrev
		}
	}
	summary := models.GoalSummary{
		TeamCode:     team,
		Scorer:       playerName(g.Name, g.FirstName, g.LastName),
		ScorerGoals:  int(g.GoalsToDate),
		Assists:      assists,
		Period:       pd.Number,
		PeriodType:   mapPeriodType(pd.PeriodType),
		TimeInPeriod: g.TimeInPeriod,
		Strength:     strings.ToUpper(g.Strength),
		EmptyNet:     g.GoalModifier == "empty-net",
		ShotType:     g.ShotType,
		HomeScore:    int(g.HomeScore),
		AwayScore:    int(g.AwayScore),
	}
	if g.EventID != 0 {
		summary.EventId = strconv.Itoa(int(g.EventID))
	}
	if g.PlayerID != 0 {
		summary.ScorerId = strconv.Itoa(int(g.PlayerID))
	}
	return summary
}

// playerName prefers the full first and last name and falls back to the
// abbreviated "K. Connor" form.
func playerName(short, first, last nhl.LandingName) string {
	if first.Default != "" && last.Default != "" {
		return first.Default + " " + last.Default
	}
	return short.Default
}

func mapPeriodType(t string) string {
	switch t {
	case "OT":
		return "OVERTIME"
	case "SO":
		return "SHOOTOUT"
	default:
		return "REGULAR"
	}
}

// applyRightRail fills the team game stats and by-period breakdowns from the
// right-rail response. When the panel is missing (fetch failed, or the game
// has not started), the previous poll's values are kept so Home Assistant
// does not see every stat drop to zero for one poll.
func applyRightRail(state *models.GameState, prev models.GameState, rr nhl.NHLRightRailResponse) {
	away, home := rr.TeamStats()
	warnUnknownStats(append(away.UnknownCategories, home.UnknownCategories...))
	mergeTeamStats(&state.Home, prev.Home, home)
	mergeTeamStats(&state.Away, prev.Away, away)

	if len(rr.Linescore.ByPeriod) > 0 {
		state.Home.PeriodScores, state.Away.PeriodScores = splitPeriods(rr.Linescore.ByPeriod)
	} else {
		state.Home.PeriodScores, state.Away.PeriodScores = prev.Home.PeriodScores, prev.Away.PeriodScores
	}
	if len(rr.ShotsByPeriod) > 0 {
		state.Home.PeriodShots, state.Away.PeriodShots = splitPeriods(rr.ShotsByPeriod)
	} else {
		state.Home.PeriodShots, state.Away.PeriodShots = prev.Home.PeriodShots, prev.Away.PeriodShots
	}
}

var (
	unknownStatMu   sync.Mutex
	unknownStatSeen = map[string]bool{}
)

// warnUnknownStats logs each unrecognised right-rail category once per
// process, so an upstream rename is visible without flooding the log.
func warnUnknownStats(categories []string) {
	unknownStatMu.Lock()
	defer unknownStatMu.Unlock()
	for _, c := range categories {
		if !unknownStatSeen[c] {
			unknownStatSeen[c] = true
			logger.Warn(fmt.Sprintf("NHL right-rail: ignoring unknown teamGameStats category %q", c))
		}
	}
}

func mergeTeamStats(team *models.TeamState, prev models.TeamState, s nhl.NHLTeamGameStats) {
	shots := team.Statistics.Shots // from landing, always fresher
	if !s.Present {
		team.Statistics = prev.Statistics
		team.Statistics.Shots = shots
		return
	}
	st := &team.Statistics
	st.HasTeamGameStats = true
	st.Hits = s.Hits
	st.Faceoffs = s.FaceoffWins
	st.FaceoffTotal = s.FaceoffTotal
	st.FaceoffPct = s.FaceoffPct
	st.PowerPlays = s.PowerPlayOpps
	st.PowerPlayGoals = s.PowerPlayGoals
	st.PowerPlayPct = s.PowerPlayPct
	st.PenaltyMinutes = s.PenaltyMinutes
	st.BlockedShots = s.BlockedShots
	st.Giveaways = s.Giveaways
	st.Takeaways = s.Takeaways
	if shots == 0 {
		st.Shots = s.ShotsOnGoal
	}
}

func splitPeriods(periods []nhl.NHLPeriodPair) (home []int, away []int) {
	home = make([]int, 0, len(periods))
	away = make([]int, 0, len(periods))
	for _, p := range periods {
		home = append(home, int(p.Home))
		away = append(away, int(p.Away))
	}
	return home, away
}
