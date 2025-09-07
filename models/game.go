package models

import (
	"encoding/json"
	"fmt"
	"time"
)

type Game struct {
	GameCode     string    `json:"gameCode"`
	LeagueId     League    `json:"leagueId"`
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
	Faceoffs   int `json:"faceoffs,omitempty"`
	PowerPlays int `json:"powerPlays,omitempty"`

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
