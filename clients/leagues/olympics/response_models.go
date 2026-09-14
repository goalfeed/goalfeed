package olympics

import (
	"encoding/json"
	"time"
)

// WMR API base: https://www.olympics.com/wmr-api/api/v2/
// competitionCode=OWG2026, disciplineCode=IHO, eventCode=TEAM6------------- for ice hockey

// StartListEntry represents a game/event unit from schedules/startList
type StartListEntry struct {
	PhaseCode     string    `json:"phaseCode"`
	EventUnitCode string    `json:"eventUnitCode"`
	StartDate     time.Time `json:"startDate,omitempty"`
	StartTime     string    `json:"startTime,omitempty"`
	Status        string    `json:"status,omitempty"`
	HomeTeam      string    `json:"homeTeam,omitempty"`
	AwayTeam      string    `json:"awayTeam,omitempty"`
	HomeTeamCode  string    `json:"homeTeamCode,omitempty"`
	AwayTeamCode  string    `json:"awayTeamCode,omitempty"`
	// Nested structures used by WMR API
	Participants []Participant `json:"participants,omitempty"`
	Result       *UnitResult   `json:"result,omitempty"`
}

// Participant represents a team/participant in an event unit
type Participant struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Type string `json:"type"` // e.g. "team"
}

// UnitResult holds result/score for an event unit
type UnitResult struct {
	HomeScore int `json:"homeScore,omitempty"`
	AwayScore int `json:"awayScore,omitempty"`
	ScoreHome int `json:"scoreHome,omitempty"`
	ScoreAway int `json:"scoreAway,omitempty"`
}

// StartListResponse is the response from schedules/startList (array or wrapper)
type StartListResponse []StartListEntry

// ScoresResponse holds results/scores response (all event units or single)
type ScoresResponse struct {
	EventUnits []ScoresEventUnit `json:"eventUnits,omitempty"`
	PhaseCode  string            `json:"phaseCode,omitempty"`
	UnitCode   string            `json:"eventUnitCode,omitempty"`
	HomeScore  int               `json:"homeScore,omitempty"`
	AwayScore  int               `json:"awayScore,omitempty"`
	ScoreHome  int               `json:"scoreHome,omitempty"`
	ScoreAway  int               `json:"scoreAway,omitempty"`
}

// ScoresEventUnit represents one game's score in a scores response
type ScoresEventUnit struct {
	PhaseCode     string `json:"phaseCode"`
	EventUnitCode string `json:"eventUnitCode"`
	HomeScore     int    `json:"homeScore,omitempty"`
	AwayScore     int    `json:"awayScore,omitempty"`
	ScoreHome     int    `json:"scoreHome,omitempty"`
	ScoreAway     int    `json:"scoreAway,omitempty"`
	Status        string `json:"status,omitempty"`
	Participants  []Participant `json:"participants,omitempty"`
}

// PlayByPlayEntry represents one play-by-play event
type PlayByPlayEntry struct {
	Sequence   int    `json:"sequence,omitempty"`
	Period     int    `json:"period,omitempty"`
	Time       string `json:"time,omitempty"`
	Clock      string `json:"clock,omitempty"`
	Type       string `json:"type,omitempty"`
	TeamCode   string `json:"teamCode,omitempty"`
	TeamName   string `json:"teamName,omitempty"`
	Description string `json:"description,omitempty"`
	IsGoal     bool   `json:"isGoal,omitempty"`
}

// PlayByPlayResponse is the response from playByPlay endpoint
type PlayByPlayResponse struct {
	Events []PlayByPlayEntry `json:"events,omitempty"`
	Plays  []PlayByPlayEntry `json:"plays,omitempty"`
}

// Live-game-load HAR: playByPlay and results/scores return wrapper with ResponseCode, TotalRows, Data.

// apiResponseWrapper is the common wrapper for playByPlay and results/scores (live-game-load HAR)
type apiResponseWrapper struct {
	ResponseCode int             `json:"ResponseCode"`
	Message      string          `json:"Message"`
	TotalRows    int             `json:"TotalRows"`
	Data         json.RawMessage `json:"Data"`
}

// playByPlayDataEntry is one item in playByPlay Data array (ActionCode GOAL/P/STARTP/ENDP/GK/TOUT, When, ScoreHome/ScoreAway as strings)
type playByPlayDataEntry struct {
	ActionCode    string `json:"ActionCode"`
	ActionOrder   int    `json:"ActionOrder"`
	When          string `json:"When"`
	PeriodCode    string `json:"PeriodCode"`
	ScoreHome     string `json:"ScoreHome"`
	ScoreAway     string `json:"ScoreAway"`
	HomeCompetitor *struct {
		OrganisationCode string `json:"OrganisationCode"`
		TeamName         string `json:"TeamName"`
	} `json:"HomeCompetitor"`
	AwayCompetitor *struct {
		OrganisationCode string `json:"OrganisationCode"`
		TeamName         string `json:"TeamName"`
	} `json:"AwayCompetitor"`
	ActionCompetitors *struct {
		OrganisationCode string `json:"OrganisationCode"`
	} `json:"ActionCompetitors"`
}

// scoresDataEntry is one item in results/scores Data array (single game: Home, Away, Scorers with Goals ScoreH/ScoreA)
type scoresDataEntry struct {
	PhaseCode     string `json:"PhaseCode"`
	EventUnitCode string `json:"EventUnitCode"`
	Home          string `json:"Home"`
	Away          string `json:"Away"`
	Scorers       []struct {
		Goals []struct {
			Order  int    `json:"Order"`
			ScoreH string `json:"ScoreH"`
			ScoreA string `json:"ScoreA"`
		} `json:"Goals"`
	} `json:"Scorers"`
}
