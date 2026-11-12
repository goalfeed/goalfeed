package models

import (
	"encoding/json"
	"fmt"
	"time"
)

type Game struct {
	GameCode     string    `json:"gameCode"`
	LeagueId     League    `json:"leagueId"`
	LeagueName   string    `json:"leagueName,omitempty"` // Display name so UI can show e.g. "Olympic Women's Hockey"
	CurrentState GameState `json:"currentState"`
	IsFetching   bool      `json:"isFetching"`
	ExtTimestamp string    `json:"extTimestamp"`
	// Enhanced fields
	GameDetails GameDetails `json:"gameDetails,omitempty"`
	Statistics  GameStats   `json:"statistics,omitempty"`
	Events      []GameEvent `json:"events,omitempty"`
}

func (g Game) GetGameKey() string {
	return fmt.Sprintf("%d-%s", g.LeagueId, g.GameCode)
}

type TeamState struct {
	Team  Team `json:"team"`
	Score int  `json:"score"`
	// Enhanced team state
	PeriodScores []int     `json:"periodScores,omitempty"`
	PeriodShots  []int     `json:"periodShots,omitempty"` // NHL: shots on goal by period
	Statistics   TeamStats `json:"statistics,omitempty"`
}

// GameState is a reflection of a games state. It contains the score and status
type GameState struct {
	Home         TeamState  `json:"home"`
	Away         TeamState  `json:"away"`
	Status       GameStatus `json:"status"`
	FetchedAt    time.Time  `json:"fetchedAt"`
	ExtTimestamp string     `json:"extTimestamp,omitempty"`
	// Enhanced game state
	Period        int     `json:"period,omitempty"`
	PeriodType    string  `json:"periodType,omitempty"` // "REGULAR", "OVERTIME", "SHOOTOUT", etc.
	TimeRemaining string  `json:"timeRemaining,omitempty"`
	Clock         string  `json:"clock,omitempty"`
	Venue         Venue   `json:"venue,omitempty"`
	Weather       Weather `json:"weather,omitempty"`
	// Baseball-specific details
	Details    EventDetails `json:"details,omitempty"`
	Statistics TeamStats    `json:"statistics,omitempty"`
	// Hockey live situation (NHL). Situation is nil when unknown, and the
	// zero value of a non-nil Situation means even strength.
	InIntermission bool          `json:"inIntermission,omitempty"`
	Situation      *Situation    `json:"situation,omitempty"`
	LastGoal       *GoalSummary  `json:"lastGoal,omitempty"`
	Goals          []GoalSummary `json:"goals,omitempty"` // every goal so far, in order
}

// GoalFor returns the goal that took teamCode's score to teamScore, if the
// upstream summary has it yet. Score detection fires one event per point, so
// this is how each event is matched to its own scorer when two goals land in
// one poll.
func (s GameState) GoalFor(teamCode string, teamScore int) (GoalSummary, bool) {
	isHome := teamCode == s.Home.Team.TeamCode
	for _, g := range s.Goals {
		if g.TeamCode != teamCode {
			continue
		}
		score := g.AwayScore
		if isHome {
			score = g.HomeScore
		}
		if score == teamScore {
			return g, true
		}
	}
	return GoalSummary{}, false
}

// Situation is the on-ice strength of a hockey game at the time of the poll.
type Situation struct {
	Code          string `json:"code,omitempty"` // NHL situationCode, e.g. "1451"
	HomeSkaters   int    `json:"homeSkaters"`
	AwaySkaters   int    `json:"awaySkaters"`
	HomeEmptyNet  bool   `json:"homeEmptyNet"` // home goalie pulled
	AwayEmptyNet  bool   `json:"awayEmptyNet"`
	PowerPlayTeam string `json:"powerPlayTeam,omitempty"` // team code with the man advantage, "" at even strength
	TimeRemaining string `json:"timeRemaining,omitempty"` // power play clock, when given
}

// Strength renders the situation from one team's side, e.g. "5v4".
func (s Situation) Strength(isHome bool) string {
	if isHome {
		return fmt.Sprintf("%dv%d", s.HomeSkaters, s.AwaySkaters)
	}
	return fmt.Sprintf("%dv%d", s.AwaySkaters, s.HomeSkaters)
}

// GoalSummary describes the most recent goal of a game as reported upstream.
type GoalSummary struct {
	EventId      string   `json:"eventId,omitempty"`
	TeamCode     string   `json:"teamCode"`
	Scorer       string   `json:"scorer"`
	ScorerId     string   `json:"scorerId,omitempty"`
	ScorerGoals  int      `json:"scorerGoals,omitempty"` // season total including this goal
	Assists      []string `json:"assists"`
	Period       int      `json:"period"`
	PeriodType   string   `json:"periodType,omitempty"`
	TimeInPeriod string   `json:"timeInPeriod"`
	Strength     string   `json:"strength"` // "EV", "PP", "SH"
	EmptyNet     bool     `json:"emptyNet"`
	ShotType     string   `json:"shotType,omitempty"`
	HomeScore    int      `json:"homeScore"`
	AwayScore    int      `json:"awayScore"`
}

type GameDetails struct {
	GameId       string        `json:"gameId"`
	Season       string        `json:"season"`
	SeasonType   string        `json:"seasonType"`
	Week         int           `json:"week,omitempty"`
	GameDate     time.Time     `json:"gameDate"`
	GameTime     string        `json:"gameTime"`
	Timezone     string        `json:"timezone"`
	Broadcasters []Broadcaster `json:"broadcasters,omitempty"`
	Officials    []Official    `json:"officials,omitempty"`
}

type Venue struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	City     string `json:"city"`
	State    string `json:"state"`
	Country  string `json:"country"`
	Capacity int    `json:"capacity,omitempty"`
	Surface  string `json:"surface,omitempty"`
	Indoor   bool   `json:"indoor,omitempty"`
}

type Weather struct {
	Temperature   int    `json:"temperature,omitempty"`
	Condition     string `json:"condition,omitempty"`
	Humidity      int    `json:"humidity,omitempty"`
	WindSpeed     int    `json:"windSpeed,omitempty"`
	WindDirection string `json:"windDirection,omitempty"`
}

type Broadcaster struct {
	Name     string `json:"name"`
	Network  string `json:"network"`
	Language string `json:"language"`
}

type Official struct {
	Name     string `json:"name"`
	Position string `json:"position"`
	Number   int    `json:"number,omitempty"`
}

type GameStats struct {
	TotalPlays       int    `json:"totalPlays"`
	TotalYards       int    `json:"totalYards,omitempty"`
	TimeOfPossession string `json:"timeOfPossession,omitempty"`
	Turnovers        int    `json:"turnovers"`
	Penalties        int    `json:"penalties"`
	PenaltyYards     int    `json:"penaltyYards,omitempty"`
}

type TeamStats struct {
	// General stats
	Plays            int    `json:"plays"`
	Yards            int    `json:"yards,omitempty"`
	TimeOfPossession string `json:"timeOfPossession,omitempty"`
	Turnovers        int    `json:"turnovers"`
	Penalties        int    `json:"penalties"`
	PenaltyYards     int    `json:"penaltyYards,omitempty"`

	// League-specific stats
	// NFL/CFL
	FirstDowns   int `json:"firstDowns,omitempty"`
	RushingYards int `json:"rushingYards,omitempty"`
	PassingYards int `json:"passingYards,omitempty"`

	// NHL
	Shots      int `json:"shots,omitempty"`
	Hits       int `json:"hits,omitempty"`
	Faceoffs   int `json:"faceoffs,omitempty"`   // faceoffs won
	PowerPlays int `json:"powerPlays,omitempty"` // power play opportunities
	// NHL right-rail team stats. HasTeamGameStats is true once the panel has
	// been read, so a zero can be told apart from "not fetched yet".
	HasTeamGameStats bool    `json:"hasTeamGameStats,omitempty"`
	FaceoffTotal     int     `json:"faceoffTotal,omitempty"`
	FaceoffPct       float64 `json:"faceoffPct,omitempty"`
	PowerPlayGoals   int     `json:"powerPlayGoals,omitempty"`
	PowerPlayPct     float64 `json:"powerPlayPct,omitempty"`
	PenaltyMinutes   int     `json:"penaltyMinutes,omitempty"`
	BlockedShots     int     `json:"blockedShots,omitempty"`
	Giveaways        int     `json:"giveaways,omitempty"`
	Takeaways        int     `json:"takeaways,omitempty"`

	// MLB
	BaseballHits int `json:"baseballHits,omitempty"`
	Errors       int `json:"errors,omitempty"`
	Strikeouts   int `json:"strikeouts,omitempty"`
	Walks        int `json:"walks,omitempty"`
}

type GameEvent struct {
	Id          string       `json:"id"`
	Type        EventType    `json:"type"`
	Period      int          `json:"period"`
	Time        string       `json:"time"`
	Clock       string       `json:"clock,omitempty"`
	Description string       `json:"description"`
	Team        Team         `json:"team"`
	Player      Player       `json:"player,omitempty"`
	Details     EventDetails `json:"details,omitempty"`
	Timestamp   time.Time    `json:"timestamp"`
}

type EventType string

const (
	EventTypeGoal         EventType = "goal"
	EventTypeAssist       EventType = "assist"
	EventTypePenalty      EventType = "penalty"
	EventTypePowerPlay    EventType = "power_play"
	EventTypeShot         EventType = "shot"
	EventTypeHit          EventType = "hit"
	EventTypeFaceoff      EventType = "faceoff"
	EventTypeSave         EventType = "save"
	EventTypeTurnover     EventType = "turnover"
	EventTypeFumble       EventType = "fumble"
	EventTypeInterception EventType = "interception"
	EventTypeTouchdown    EventType = "touchdown"
	EventTypeFieldGoal    EventType = "field_goal"
	EventTypeSafety       EventType = "safety"
	EventTypeHomeRun      EventType = "home_run"
	EventTypeStrikeout    EventType = "strikeout"
	EventTypeWalk         EventType = "walk"
	EventTypeError        EventType = "error"
	EventTypePeriodStart  EventType = "period_start"
	EventTypePeriodEnd    EventType = "period_end"
	EventTypeGameStart    EventType = "game_start"
	EventTypeGameEnd      EventType = "game_end"
)

type Player struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Number   int    `json:"number"`
	Position string `json:"position"`
	Team     Team   `json:"team"`
}

type BaseRunners struct {
	First  *Player `json:"first,omitempty"`
	Second *Player `json:"second,omitempty"`
	Third  *Player `json:"third,omitempty"`
}

type EventDetails struct {
	// Goal details
	GoalType string `json:"goalType,omitempty"` // "even_strength", "power_play", "short_handed"
	Assist1  Player `json:"assist1,omitempty"`
	Assist2  Player `json:"assist2,omitempty"`

	// Penalty details
	PenaltyType    string `json:"penaltyType,omitempty"`
	PenaltyMinutes int    `json:"penaltyMinutes,omitempty"`

	// Play details
	YardLine    int    `json:"yardLine,omitempty"`
	Down        int    `json:"down,omitempty"`
	Distance    int    `json:"distance,omitempty"`
	YardsGained int    `json:"yardsGained,omitempty"`
	Possession  string `json:"possession,omitempty"` // Team code with possession

	// Baseball details
	Inning      int    `json:"inning,omitempty"`
	Outs        int    `json:"outs,omitempty"`
	Bases       string `json:"bases,omitempty"` // "empty", "1st", "2nd", "3rd", "loaded"
	PitchCount  int    `json:"pitchCount,omitempty"`
	StrikeCount int    `json:"strikeCount,omitempty"`
	BallCount   int    `json:"ballCount,omitempty"`
	// Enhanced baseball details
	BaseRunners BaseRunners `json:"baseRunners,omitempty"`
	Pitcher     Player      `json:"pitcher,omitempty"`
	Batter      Player      `json:"batter,omitempty"`
}

type GameUpdate struct {
	OldState GameState
	NewState GameState
	Events   []GameEvent
}

const (
	StatusUpcoming = iota
	StatusActive
	StatusDelayed
	StatusEnded
)

type GameStatus int

// MarshalJSON converts GameStatus to string for JSON serialization
func (gs GameStatus) MarshalJSON() ([]byte, error) {
	switch gs {
	case StatusUpcoming:
		return json.Marshal("upcoming")
	case StatusActive:
		return json.Marshal("active")
	case StatusDelayed:
		return json.Marshal("delayed")
	case StatusEnded:
		return json.Marshal("ended")
	default:
		return json.Marshal("unknown")
	}
}

// UnmarshalJSON parses GameStatus from string or number for JSON deserialization
func (gs *GameStatus) UnmarshalJSON(data []byte) error {
	// Try string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		switch s {
		case "upcoming":
			*gs = StatusUpcoming
			return nil
		case "active":
			*gs = StatusActive
			return nil
		case "delayed":
			*gs = StatusDelayed
			return nil
		case "ended":
			*gs = StatusEnded
			return nil
		default:
			// Unknown string maps to upcoming to avoid repeated starts
			*gs = StatusUpcoming
			return nil
		}
	}
	// Fallback to numeric
	var n int
	if err := json.Unmarshal(data, &n); err == nil {
		*gs = GameStatus(n)
		return nil
	}
	// Default
	*gs = StatusUpcoming
	return nil
}
