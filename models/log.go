package models

import "time"

// AppLogType enumerates different kinds of log entries stored by Goalfeed
type AppLogType string

const (
	AppLogTypeEvent       AppLogType = "event"
	AppLogTypeStateChange AppLogType = "state_change"
	AppLogTypeLogLine     AppLogType = "log"
)

// AppLogLevel represents severity for generic log lines
type AppLogLevel string

const (
	AppLogLevelDebug AppLogLevel = "debug"
	AppLogLevelInfo  AppLogLevel = "info"
	AppLogLevelWarn  AppLogLevel = "warn"
	AppLogLevelError AppLogLevel = "error"
)

// AppLogEntry captures events and state changes for UI history and diagnostics
type AppLogEntry struct {
	Id            string      `json:"id"`
	Type          AppLogType  `json:"type"`
	Level         AppLogLevel `json:"level,omitempty"`
	LeagueId      League      `json:"leagueId"`
	LeagueName    string      `json:"leagueName"`
	TeamCode      string      `json:"teamCode"`
	Opponent      string      `json:"opponent,omitempty"`
	GameCode      string      `json:"gameCode,omitempty"`
	Metric        string      `json:"metric,omitempty"` // for state_change
	Before        interface{} `json:"before,omitempty"` // previous value
	After         interface{} `json:"after,omitempty"`  // new value
	Event         *Event      `json:"event,omitempty"`  // for event type
	Message       string      `json:"message,omitempty"`
	Source        string      `json:"source,omitempty"`
	Target        string      `json:"target,omitempty"`  // e.g., HA entity_id or event name
	Success       *bool       `json:"success,omitempty"` // delivery result, if applicable
	Error         string      `json:"error,omitempty"`
	CorrelationId string      `json:"correlationId,omitempty"` // link to event.id
	Timestamp     time.Time   `json:"timestamp"`
}
