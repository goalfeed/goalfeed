package nhl

import (
	"time"
)

type NHLScoreboardResponse struct {
	ID                int             `json:"id,omitempty"`
	Season            int             `json:"season,omitempty"`
	GameType          int             `json:"gameType,omitempty"`
	GameDate          string          `json:"gameDate,omitempty"`
	Venue             Venue           `json:"venue,omitempty"`
	StartTimeUTC      time.Time       `json:"startTimeUTC,omitempty"`
	EasternUTCOffset  string          `json:"easternUTCOffset,omitempty"`
	VenueUTCOffset    string          `json:"venueUTCOffset,omitempty"`
	VenueTimezone     string          `json:"venueTimezone,omitempty"`
	GameState         string          `json:"gameState,omitempty"`
	GameScheduleState string          `json:"gameScheduleState,omitempty"`
	AwayTeam          NHLScheduleTeam `json:"awayTeam,omitempty"`
	HomeTeam          NHLScheduleTeam `json:"homeTeam,omitempty"`
	ShootoutInUse     bool            `json:"shootoutInUse,omitempty"`
	MaxPeriods        int             `json:"maxPeriods,omitempty"`
	RegPeriods        int             `json:"regPeriods,omitempty"`
	OtInUse           bool            `json:"otInUse,omitempty"`
	TiesInUse         bool            `json:"tiesInUse,omitempty"`
	TicketsLink       string          `json:"ticketsLink,omitempty"`
}
