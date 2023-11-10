package models

import (
	"fmt"
	"time"
)

type Game struct {
	GameCode     string    `json:"GameCode"`
	LeagueId     League    `json:"LeagueId"`
	CurrentState GameState `json:"CurrentState"`
	IsFetching   bool      `json:"IsFetching"`
	ExtTimestamp string    `json:"ExtTimestamp"`
}

func (g Game) GetGameKey() string {
	return fmt.Sprintf("%d-%s", g.LeagueId, g.GameCode)
}

type TeamState struct {
	Team  Team `json:"Team"`
	Score int  `json:"Score"`
}

// GameState is a reflection of a games state. It contains the score and status
type GameState struct {
	Home         TeamState  `json:"Home"`
	Away         TeamState  `json:"Away"`
	Status       GameStatus `json:"Status"`
	FetchedAt    time.Time  `json:"FetchedAt"`
	ExtTimestamp string     `json:"ExtTimestamp,omitempty"`
}

type GameUpdate struct {
	OldState GameState
	NewState GameState
}

const (
	StatusUpcoming = iota
	StatusActive
	StatusEnded
)

type GameStatus int
