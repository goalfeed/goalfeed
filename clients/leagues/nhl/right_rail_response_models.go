package nhl

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
)

// NHLRightRailResponse is the subset of /v1/gamecenter/{id}/right-rail that
// goalfeed reads: the team game stats panel, shots by period and the
// linescore.
//
// teamGameStats is a list of {category, awayValue, homeValue} rows whose value
// type depends on the category: "sog" and "hits" are integers,
// "faceoffWinningPctg" is a float, and "faceoffWins" and "powerPlay" are
// strings like "28/77". The values are therefore kept raw and parsed per
// category, so one surprising row can never discard the rest of the panel.
type NHLRightRailResponse struct {
	TeamGameStats []NHLTeamGameStatRow `json:"teamGameStats"`
	ShotsByPeriod []NHLPeriodPair      `json:"shotsByPeriod"`
	Linescore     NHLLinescore         `json:"linescore"`
}

type NHLTeamGameStatRow struct {
	Category  string          `json:"category"`
	AwayValue json.RawMessage `json:"awayValue"`
	HomeValue json.RawMessage `json:"homeValue"`
}

type NHLPeriodPair struct {
	PeriodDescriptor PeriodDescriptor `json:"periodDescriptor"`
	Away             FlexInt          `json:"away"`
	Home             FlexInt          `json:"home"`
}

type NHLLinescore struct {
	ByPeriod []NHLPeriodPair `json:"byPeriod"`
	Totals   struct {
		Away FlexInt `json:"away"`
		Home FlexInt `json:"home"`
	} `json:"totals"`
}

// NHLTeamGameStats is one side of the right-rail team stats panel, parsed into
// typed values. Present is false when the panel was missing or empty, so a
// caller can keep the last good numbers instead of publishing zeros.
type NHLTeamGameStats struct {
	Present bool

	ShotsOnGoal       int
	FaceoffWins       int
	FaceoffTotal      int
	FaceoffPct        float64 // 0-100, one decimal place
	PowerPlayGoals    int
	PowerPlayOpps     int
	PowerPlayPct      float64 // 0-100, one decimal place
	PenaltyMinutes    int
	Hits              int
	BlockedShots      int
	Giveaways         int
	Takeaways         int
	UnknownCategories []string
}

// TeamStats parses the team game stats panel into away and home values.
// Categories goalfeed does not know are recorded in UnknownCategories rather
// than dropped silently, so a renamed upstream field shows up in logs.
func (r NHLRightRailResponse) TeamStats() (away NHLTeamGameStats, home NHLTeamGameStats) {
	if len(r.TeamGameStats) == 0 {
		return away, home
	}
	away.Present, home.Present = true, true
	for _, row := range r.TeamGameStats {
		if !applyStat(&away, row.Category, row.AwayValue) {
			away.UnknownCategories = append(away.UnknownCategories, row.Category)
		}
		if !applyStat(&home, row.Category, row.HomeValue) {
			home.UnknownCategories = append(home.UnknownCategories, row.Category)
		}
	}
	return away, home
}

func applyStat(s *NHLTeamGameStats, category string, raw json.RawMessage) bool {
	switch category {
	case "sog":
		s.ShotsOnGoal = rawInt(raw)
	case "faceoffWins":
		s.FaceoffWins, s.FaceoffTotal = rawRatio(raw)
	case "faceoffWinningPctg":
		s.FaceoffPct = rawPct(raw)
	case "powerPlay":
		s.PowerPlayGoals, s.PowerPlayOpps = rawRatio(raw)
	case "powerPlayPctg":
		s.PowerPlayPct = rawPct(raw)
	case "pim":
		s.PenaltyMinutes = rawInt(raw)
	case "hits":
		s.Hits = rawInt(raw)
	case "blockedShots":
		s.BlockedShots = rawInt(raw)
	case "giveaways":
		s.Giveaways = rawInt(raw)
	case "takeaways":
		s.Takeaways = rawInt(raw)
	default:
		return false
	}
	return true
}

// rawNumber reads a JSON number or a numeric string. Anything else is 0.
func rawNumber(raw json.RawMessage) float64 {
	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		return f
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if v, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err == nil {
			return v
		}
	}
	return 0
}

func rawInt(raw json.RawMessage) int {
	return int(math.Round(rawNumber(raw)))
}

// rawPct turns a 0-1 fraction into a 0-100 percentage rounded to one place.
// A value already above 1 is assumed to be a percentage.
func rawPct(raw json.RawMessage) float64 {
	v := rawNumber(raw)
	if v <= 1 {
		v *= 100
	}
	return math.Round(v*10) / 10
}

// rawRatio reads "made/attempted" strings such as "1/3" or "28/77".
func rawRatio(raw json.RawMessage) (int, int) {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return 0, 0
	}
	a, b, ok := strings.Cut(strings.TrimSpace(s), "/")
	if !ok {
		return 0, 0
	}
	made, err1 := strconv.Atoi(strings.TrimSpace(a))
	total, err2 := strconv.Atoi(strings.TrimSpace(b))
	if err1 != nil || err2 != nil {
		return 0, 0
	}
	return made, total
}

// FlexInt accepts an integer encoded either as a JSON number or as a numeric
// string, the inconsistency AGENTS.md warns about. A non-numeric value decodes
// as 0 rather than failing the whole response.
type FlexInt int

func (f *FlexInt) UnmarshalJSON(data []byte) error {
	*f = FlexInt(rawInt(data))
	return nil
}
