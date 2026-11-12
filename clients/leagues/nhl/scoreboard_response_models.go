package nhl

import (
	"time"
)

type Clock struct {
	TimeRemaining    string `json:"timeRemaining,omitempty"`
	SecondsRemaining int    `json:"secondsRemaining,omitempty"`
	Running          bool   `json:"running,omitempty"`
	InIntermission   bool   `json:"inIntermission,omitempty"`
}

type NHLScoreboardResponse struct {
	ID                int              `json:"id,omitempty"`
	Season            int              `json:"season,omitempty"`
	GameType          int              `json:"gameType,omitempty"`
	GameDate          string           `json:"gameDate,omitempty"`
	Venue             Venue            `json:"venue,omitempty"`
	StartTimeUTC      time.Time        `json:"startTimeUTC,omitempty"`
	EasternUTCOffset  string           `json:"easternUTCOffset,omitempty"`
	VenueUTCOffset    string           `json:"venueUTCOffset,omitempty"`
	VenueTimezone     string           `json:"venueTimezone,omitempty"`
	GameState         string           `json:"gameState,omitempty"`
	GameScheduleState string           `json:"gameScheduleState,omitempty"`
	AwayTeam          NHLScheduleTeam  `json:"awayTeam,omitempty"`
	HomeTeam          NHLScheduleTeam  `json:"homeTeam,omitempty"`
	ShootoutInUse     bool             `json:"shootoutInUse,omitempty"`
	MaxPeriods        int              `json:"maxPeriods,omitempty"`
	RegPeriods        int              `json:"regPeriods,omitempty"`
	OtInUse           bool             `json:"otInUse,omitempty"`
	TiesInUse         bool             `json:"tiesInUse,omitempty"`
	TicketsLink       string           `json:"ticketsLink,omitempty"`
	PeriodDescriptor  PeriodDescriptor `json:"periodDescriptor,omitempty"`
	Clock             Clock            `json:"clock,omitempty"`
	// Situation is present only while a game is live and not at plain
	// 5-on-5 with both goalies in; nil means even strength.
	Situation *LandingSituation `json:"situation,omitempty"`
	Summary   LandingSummary    `json:"summary,omitempty"`
}

// LandingSituation is the landing page's live strength block, e.g. a home
// power play: {"homeTeam":{"abbrev":"WPG","situationDescriptions":["PP"],
// "strength":5},"awayTeam":{"abbrev":"NJD","strength":4},
// "situationCode":"1451","timeRemaining":"01:46"}.
type LandingSituation struct {
	HomeTeam      LandingSituationTeam `json:"homeTeam"`
	AwayTeam      LandingSituationTeam `json:"awayTeam"`
	SituationCode string               `json:"situationCode"`
	TimeRemaining string               `json:"timeRemaining"`
}

type LandingSituationTeam struct {
	Abbrev                string   `json:"abbrev"`
	SituationDescriptions []string `json:"situationDescriptions"`
	Strength              FlexInt  `json:"strength"`
}

type LandingSummary struct {
	Scoring []LandingScoringPeriod `json:"scoring"`
}

type LandingScoringPeriod struct {
	PeriodDescriptor PeriodDescriptor `json:"periodDescriptor"`
	Goals            []LandingGoal    `json:"goals"`
}

type LandingName struct {
	Default string `json:"default"`
}

type LandingGoal struct {
	EventID       FlexInt         `json:"eventId"`
	SituationCode string          `json:"situationCode"`
	Strength      string          `json:"strength"`     // "ev", "pp", "sh"
	GoalModifier  string          `json:"goalModifier"` // "none", "empty-net", "penalty-shot", ...
	PlayerID      FlexInt         `json:"playerId"`
	Name          LandingName     `json:"name"`
	FirstName     LandingName     `json:"firstName"`
	LastName      LandingName     `json:"lastName"`
	TeamAbbrev    LandingName     `json:"teamAbbrev"`
	GoalsToDate   FlexInt         `json:"goalsToDate"`
	AwayScore     FlexInt         `json:"awayScore"`
	HomeScore     FlexInt         `json:"homeScore"`
	TimeInPeriod  string          `json:"timeInPeriod"`
	ShotType      string          `json:"shotType"`
	IsHome        bool            `json:"isHome"`
	Assists       []LandingAssist `json:"assists"`
}

type LandingAssist struct {
	PlayerID      FlexInt     `json:"playerId"`
	Name          LandingName `json:"name"`
	FirstName     LandingName `json:"firstName"`
	LastName      LandingName `json:"lastName"`
	AssistsToDate FlexInt     `json:"assistsToDate"`
}
