package models

import "time"

type Event struct {
	// Basic event info
	Id          string    `json:"id"`
	Type        EventType `json:"type"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`

	// Team and player info
	TeamCode     string `json:"teamCode"`
	TeamName     string `json:"teamName"`
	TeamHash     string `json:"teamHash"`
	PlayerName   string `json:"playerName,omitempty"`
	PlayerNumber int    `json:"playerNumber,omitempty"`

	// Game context
	LeagueId   int    `json:"leagueId"`
	LeagueName string `json:"leagueName"`
	GameCode   string `json:"gameCode"`
	GameId     string `json:"gameId"`
	Period     int    `json:"period"`
	Time       string `json:"time"`
	Clock      string `json:"clock,omitempty"`

	// Opponent info
	OpponentCode string `json:"opponentCode"`
	OpponentName string `json:"opponentName"`
	OpponentHash string `json:"opponentHash"`

	// Event details
	Details EventDetails `json:"details,omitempty"`
	Score   ScoreUpdate  `json:"score,omitempty"`

	// Venue and broadcast info
	Venue        Venue         `json:"venue,omitempty"`
	Broadcasters []Broadcaster `json:"broadcasters,omitempty"`
}

type ScoreUpdate struct {
	HomeScore int    `json:"homeScore"`
	AwayScore int    `json:"awayScore"`
	HomeTeam  string `json:"homeTeam"`
	AwayTeam  string `json:"awayTeam"`
}

// Enhanced event for Home Assistant and other consumers
type RichEvent struct {
	Event
	// Additional context for external systems
	GameState     GameState     `json:"gameState"`
	TeamStats     TeamStats     `json:"teamStats,omitempty"`
	PlayerStats   PlayerStats   `json:"playerStats,omitempty"`
	Weather       Weather       `json:"weather,omitempty"`
	BroadcastInfo BroadcastInfo `json:"broadcastInfo,omitempty"`
}

type PlayerStats struct {
	Player         Player `json:"player"`
	Goals          int    `json:"goals,omitempty"`
	Assists        int    `json:"assists,omitempty"`
	Points         int    `json:"points,omitempty"`
	Shots          int    `json:"shots,omitempty"`
	Hits           int    `json:"hits,omitempty"`
	PenaltyMinutes int    `json:"penaltyMinutes,omitempty"`
	// Football stats
	RushingYards int `json:"rushingYards,omitempty"`
	PassingYards int `json:"passingYards,omitempty"`
	Receptions   int `json:"receptions,omitempty"`
	Touchdowns   int `json:"touchdowns,omitempty"`
	// Baseball stats
	BaseballHits int `json:"baseballHits,omitempty"`
	RBIs         int `json:"rbis,omitempty"`
	Strikeouts   int `json:"strikeouts,omitempty"`
	Walks        int `json:"walks,omitempty"`
}

type BroadcastInfo struct {
	Networks     []string `json:"networks"`
	Streaming    []string `json:"streaming,omitempty"`
	Radio        []string `json:"radio,omitempty"`
	Language     string   `json:"language"`
	Availability string   `json:"availability"` // "free", "premium", "regional"
}

// Event priority for Home Assistant notifications
type EventPriority int

const (
	PriorityLow EventPriority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

func (ep EventPriority) String() string {
	switch ep {
	case PriorityLow:
		return "low"
	case PriorityNormal:
		return "normal"
	case PriorityHigh:
		return "high"
	case PriorityCritical:
		return "critical"
	default:
		return "normal"
	}
}

// GetEventPriority determines the priority based on event type
func (e Event) GetEventPriority() EventPriority {
	switch e.Type {
	case EventTypeGoal, EventTypeTouchdown, EventTypeHomeRun:
		return PriorityHigh
	case EventTypeGameStart, EventTypeGameEnd, EventTypePeriodStart, EventTypePeriodEnd:
		return PriorityNormal
	case EventTypePenalty, EventTypeTurnover, EventTypeFumble:
		return PriorityHigh
	default:
		return PriorityNormal
	}
}

// GetEventIcon returns an appropriate icon for the event type
func (e Event) GetEventIcon() string {
	switch e.Type {
	case EventTypeGoal:
		return "🏒"
	case EventTypeTouchdown:
		return "🏈"
	case EventTypeHomeRun:
		return "⚾"
	case EventTypePenalty:
		return "⚠️"
	case EventTypePowerPlay:
		return "⚡"
	case EventTypeShot:
		return "🎯"
	case EventTypeSave:
		return "🛡️"
	case EventTypeStrikeout:
		return "⚡"
	case EventTypeWalk:
		return "🚶"
	case EventTypeError:
		return "❌"
	case EventTypeGameStart:
		return "🏁"
	case EventTypeGameEnd:
		return "🏁"
	case EventTypePeriodStart:
		return "⏰"
	case EventTypePeriodEnd:
		return "⏰"
	default:
		return "📰"
	}
}

// GetEventColor returns a color for the event type
func (e Event) GetEventColor() string {
	switch e.Type {
	case EventTypeGoal, EventTypeTouchdown, EventTypeHomeRun:
		return "green"
	case EventTypePenalty, EventTypeTurnover, EventTypeFumble, EventTypeError:
		return "red"
	case EventTypePowerPlay, EventTypeStrikeout:
		return "yellow"
	case EventTypeGameStart, EventTypeGameEnd:
		return "blue"
	case EventTypePeriodStart, EventTypePeriodEnd:
		return "purple"
	default:
		return "gray"
	}
}
