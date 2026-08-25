package cfl

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// FlexInt unmarshals a JSON int OR a JSON string containing digits into an
// int. The BetGenius live-game feed sends some numeric-looking fields
// (sportId, competitionId) as JSON strings (e.g. "sportId":"17") even
// though they're conceptually numbers; a plain `int` field fails to
// unmarshal at all in that case, and because json.Unmarshal aborts the
// whole call on a type-mismatch error, a single such field silently
// discards the entire otherwise-valid response (see GetCFLLiveGame, which
// used to return a zero struct on any unmarshal error). FlexInt accepts
// both encodings so a future upstream flip back to real numbers keeps
// working too.
type FlexInt int

func (f *FlexInt) UnmarshalJSON(data []byte) error {
	// Numeric form: "sportId":17
	var n int
	if err := json.Unmarshal(data, &n); err == nil {
		*f = FlexInt(n)
		return nil
	}
	// String form: "sportId":"17"
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		if s == "" {
			*f = 0
			return nil
		}
		v, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("cfl: FlexInt: cannot parse %q as int: %w", s, err)
		}
		*f = FlexInt(v)
		return nil
	}
	return fmt.Errorf("cfl: FlexInt: value is neither a number nor a numeric string: %s", string(data))
}

func (f FlexInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(f))
}

// CFL Schedule Response Models
type CFLScheduleResponse []CFLRound

type CFLRound struct {
	ID          int       `json:"id"`
	Status      string    `json:"status"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Number      int       `json:"number"`
	StartDate   string    `json:"startDate"`
	EndDate     string    `json:"endDate"`
	Tournaments []CFLGame `json:"tournaments"`
}

type CFLGame struct {
	ID           int         `json:"id"`
	Date         string      `json:"date"`
	Status       string      `json:"status"`
	HomeSquad    CFLTeam     `json:"homeSquad"`
	AwaySquad    CFLTeam     `json:"awaySquad"`
	ActivePeriod interface{} `json:"activePeriod"`
	Timeouts     CFLTimeouts `json:"timeouts"`
	Possession   string      `json:"possession"`
	CFLID        int         `json:"cflId"`
	Clock        string      `json:"clock"`
	Winner       interface{} `json:"winner"`
	IsHidden     bool        `json:"isHidden"`
	Markets      interface{} `json:"markets"`
	MarketsBCLC  interface{} `json:"marketsBCLC"`
}

type CFLTeam struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	Score     int    `json:"score"`
}

type CFLTimeouts struct {
	Away int `json:"away"`
	Home int `json:"home"`
}

// CFL Live Game Response Models (from BetGenius API)
//
// SportID and CompetitionID are FlexInt (not int): a live capture of this
// endpoint showed both sent as JSON strings ("sportId":"17",
// "competitionId":"1035") even though the values are numeric. A plain int
// field fails to unmarshal on that shape, and because json.Unmarshal
// aborts the whole call on the first type-mismatch error, that alone was
// enough to make GetCFLLiveGame discard the entire response (including a
// correctly-typed score) and return a zero struct. The rest of this type
// was audited against the same live capture and found to already match
// (matchStatus/currentPhase/possession as strings, scores/timeouts/down/
// yardsToGo as numbers, clockUnreliable as bool, availableTabs as bools).
type CFLLiveGameResponse struct {
	Data          CFLLiveGameData  `json:"data"`
	Sport         string           `json:"sport"`
	SportID       FlexInt          `json:"sportId"`
	CompetitionID FlexInt          `json:"competitionId"`
	AvailableTabs CFLAvailableTabs `json:"availableTabs"`
}

type CFLLiveGameData struct {
	BetGeniusFixtureID string                `json:"betGeniusFixtureId"`
	PreferredSourceIDs CFLPreferredSourceIDs `json:"preferredSourceIds"`
	ScoreboardInfo     CFLScoreboardInfo     `json:"scoreboardInfo"`
	LiveStream         CFLLiveStream         `json:"liveStream"`
	MatchInfo          CFLMatchInfo          `json:"matchInfo"`
	Court              CFLCourt              `json:"court"`
	Scheduler          bool                  `json:"scheduler"`
}

type CFLPreferredSourceIDs struct {
	MatchActionsSourceID string `json:"matchActionsSourceId"`
	PlayerStatsSourceID  string `json:"playerStatsSourceId"`
	TeamStatsSourceID    string `json:"teamStatsSourceId"`
}

type CFLScoreboardInfo struct {
	MatchStatus          string           `json:"matchStatus"`
	CurrentPhase         string           `json:"currentPhase"`
	AwayScore            int              `json:"awayScore"`
	HomeScore            int              `json:"homeScore"`
	AwayTimeoutsLeft     int              `json:"awayTimeoutsLeft"`
	HomeTimeoutsLeft     int              `json:"homeTimeoutsLeft"`
	TotalTimeouts        int              `json:"totalTimeouts"`
	ScoreByPhases        CFLScoreByPhases `json:"scoreByPhases"`
	TimeRemainingInPhase string           `json:"timeRemainingInPhase"`
	Possession           string           `json:"possession"`
	Down                 interface{}      `json:"down"`
	YardsToGo            interface{}      `json:"yardsToGo"`
	TotalPhases          int              `json:"totalPhases"`
	PhaseQualifier       string           `json:"phaseQualifier"`
	ClockUnreliable      bool             `json:"clockUnreliable"`
}

type CFLScoreByPhases struct {
	AwayScore CFLPhaseScore `json:"awayScore"`
	HomeScore CFLPhaseScore `json:"homeScore"`
}

type CFLPhaseScore struct {
	Quarter1 int `json:"quarter1"`
}

type CFLMatchInfo struct {
	RoundID            string          `json:"roundId"`
	RoundName          string          `json:"roundName"`
	ScheduledStartTime string          `json:"scheduledStartTime"`
	VenueName          string          `json:"venueName"`
	SeasonID           string          `json:"seasonId"`
	SeasonName         string          `json:"seasonName"`
	HomeTeam           CFLDetailedTeam `json:"homeTeam"`
	AwayTeam           CFLDetailedTeam `json:"awayTeam"`
	PlayedPhases       []string        `json:"playedPhases"`
}

type CFLDetailedTeam struct {
	FullName     string         `json:"fullName"`
	CompetitorID string         `json:"competitorId"`
	Details      CFLTeamDetails `json:"details"`
}

type CFLTeamDetails struct {
	Key            string   `json:"key"`
	Brand          CFLBrand `json:"brand"`
	PrimaryColor   string   `json:"primaryColor"`
	SecondaryColor string   `json:"secondaryColor"`
	FirstName      string   `json:"firstName"`
	ShortName      string   `json:"shortName"`
	SecondName     string   `json:"secondName"`
	Abbreviation   string   `json:"abbreviation"`
	OfficialName   string   `json:"officialName"`
}

type CFLBrand struct {
	Logo  string   `json:"logo"`
	Theme CFLTheme `json:"theme"`
}

type CFLTheme struct {
	Dark  CFLThemeColors `json:"dark"`
	Light CFLThemeColors `json:"light"`
}

type CFLThemeColors struct {
	Logo           CFLThemeLogo `json:"logo"`
	PrimaryColor   string       `json:"primaryColor"`
	SecondaryColor string       `json:"secondaryColor"`
}

type CFLThemeLogo struct {
	SVG string `json:"svg"`
}

type CFLCourt struct {
	MatchActions []interface{} `json:"matchActions"`
}

type CFLAvailableTabs struct {
	Court       bool `json:"court"`
	TeamStats   bool `json:"teamStats"`
	PlayerStats bool `json:"playerStats"`
	Lineups     bool `json:"lineups"`
	PlayByPlay  bool `json:"playByPlay"`
}

// Live Stream Data Structures
type CFLLiveStream struct {
	CurrentPlay  CFLCurrentPlay  `json:"currentPlay"`
	CurrentDrive CFLCurrentDrive `json:"currentDrive"`
	Actions      []CFLAction     `json:"actions"`
}

type CFLCurrentPlay struct {
	DownNumber      int         `json:"downNumber"`
	LineOfScrimmage int         `json:"lineOfScrimmage"`
	FirstDownLine   int         `json:"firstDownLine"`
	PlayType        string      `json:"playType"`
	YardsToGo       int         `json:"yardsToGo"`
	Possession      string      `json:"possession"`
	Clock           string      `json:"clock"`
	Phase           string      `json:"phase"`
	PlayFormation   string      `json:"playFormation,omitempty"`
	Quarterback     int         `json:"quarterback,omitempty"`
	YardLine        CFLYardLine `json:"yardLine,omitempty"`
}

type CFLYardLine struct {
	TeamNumber int `json:"teamNumber"`
	YardLine   int `json:"yardLine"`
}

type CFLCurrentDrive struct {
	YardsGained           int         `json:"yardsGained"`
	TimeOfPossession      string      `json:"timeOfPossession"`
	HowObtained           string      `json:"howObtained"`
	HowLost               string      `json:"howLost"`
	Plays                 int         `json:"plays"`
	StartingFieldPosition int         `json:"startingFieldPosition"`
	CurrentFieldPosition  int         `json:"currentFieldPosition"`
	DriveStart            CFLYardLine `json:"driveStart,omitempty"`
	DriveEnd              CFLYardLine `json:"driveEnd,omitempty"`
}

type CFLAction struct {
	ActionType  string `json:"actionType"`
	Description string `json:"description"`
	YardsGained int    `json:"yardsGained"`
	Clock       string `json:"clock"`
	Down        int    `json:"down"`
	Distance    int    `json:"distance"`
	Possession  string `json:"possession"`
	YardLine    int    `json:"yardLine,omitempty"`
	PlayerId    int    `json:"playerId,omitempty"`
	ReceiverId  int    `json:"receiverId,omitempty"`
}
