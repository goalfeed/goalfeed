package iihf

import (
	"time"
)

type IIHFScheduleResponseGame struct {
	GuestTeam       IIHFScheduleTeam `json:"GuestTeam"`
	HomeTeam        IIHFScheduleTeam `json:"HomeTeam"`
	GameDateTimeUTC time.Time        `json:"GameDateTimeUTC"`
	EventStatus     string           `json:"EventStatus"`
	Status          string           `json:"Status"`
	GameNumber      string           `json:"GameNumber"`
	GameID          string           `json:"GameId"`
}

type IIHFScheduleTeam struct {
	Points   int    `json:"Points"`
	TeamCode string `json:"TeamCode"`
}

type IIHFScheduleResponse []IIHFScheduleResponseGame

type IIHFGameScoreResponse struct {
	GameID          string `json:"GameId"`
	GameNumber      string `json:"GameNumber"`
	EventID         string `json:"EventId"`
	Status          string `json:"Status"`
	IsGameCompleted bool   `json:"IsGameCompleted"`
	HomeTeam        struct {
		ShortTeamName string `json:"ShortTeamName"`
		LongTeamName  string `json:"LongTeamName"`
		Color         string `json:"Color"`
	} `json:"HomeTeam"`
	AwayTeam struct {
		ShortTeamName string `json:"ShortTeamName"`
		LongTeamName  string `json:"LongTeamName"`
		Color         string `json:"Color"`
	} `json:"AwayTeam"`
	CurrentScore struct {
		Home int `json:"Home"`
		Away int `json:"Away"`
	} `json:"CurrentScore"`
}
