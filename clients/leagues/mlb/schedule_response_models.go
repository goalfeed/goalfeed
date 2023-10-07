package mlb

import (
	"time"
)

type MLBScheduleResponse struct {
	Copyright            string  `json:"copyright"`
	Totalitems           int     `json:"totalItems"`
	Totalevents          int     `json:"totalEvents"`
	Totalgames           int     `json:"totalGames"`
	Totalgamesinprogress int     `json:"totalGamesInProgress"`
	Dates                []Dates `json:"dates"`
}
type LeagueRecord struct {
	Wins   int    `json:"wins"`
	Losses int    `json:"losses"`
	Pct    string `json:"pct"`
}
type TeamInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Link string `json:"link"`
}
type MLBScheduleTeam struct {
	LeagueRecord LeagueRecord `json:"leagueRecord"`
	Score        int          `json:"score"`
	Team         TeamInfo     `json:"team"`
	Splitsquad   bool         `json:"splitSquad"`
	Seriesnumber int          `json:"seriesNumber"`
}
type Teams struct {
	Away MLBScheduleTeam `json:"away"`
	Home MLBScheduleTeam `json:"home"`
}
type Venue struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Link string `json:"link"`
}
type Content struct {
	Link string `json:"link"`
}
type Status struct {
	AbstractGameState string `json:"abstractGameState"`
	CodedGameState    string `json:"codedGameState"`
	DetailedState     string `json:"detailedState"`
	StatusCode        string `json:"statusCode"`
	StartTimetbd      bool   `json:"startTimeTBD"`
	Reason            string `json:"reason"`
	AbstractGameCode  string `json:"abstractGameCode"`
}
type MLBScheduleResponseGame struct {
	GamePk                 int       `json:"gamePk"`
	Link                   string    `json:"link"`
	Gametype               string    `json:"gameType"`
	Season                 string    `json:"season"`
	Gamedate               time.Time `json:"gameDate"`
	Officialdate           string    `json:"officialDate"`
	Status                 Status    `json:"status,omitempty"`
	Teams                  Teams     `json:"teams"`
	Venue                  Venue     `json:"venue"`
	Content                Content   `json:"content"`
	Gamenumber             int       `json:"gameNumber"`
	Publicfacing           bool      `json:"publicFacing"`
	Doubleheader           string    `json:"doubleHeader"`
	Gamedaytype            string    `json:"gamedayType"`
	Tiebreaker             string    `json:"tiebreaker"`
	Calendareventid        string    `json:"calendarEventID"`
	Seasondisplay          string    `json:"seasonDisplay"`
	Daynight               string    `json:"dayNight"`
	Scheduledinnings       int       `json:"scheduledInnings"`
	Reversehomeawaystatus  bool      `json:"reverseHomeAwayStatus"`
	Inningbreaklength      int       `json:"inningBreakLength"`
	Gamesinseries          int       `json:"gamesInSeries"`
	Seriesgamenumber       int       `json:"seriesGameNumber"`
	Seriesdescription      string    `json:"seriesDescription"`
	Recordsource           string    `json:"recordSource"`
	Ifnecessary            string    `json:"ifNecessary"`
	Ifnecessarydescription string    `json:"ifNecessaryDescription"`
}
type Dates struct {
	Date                 string                    `json:"date"`
	Totalitems           int                       `json:"totalItems"`
	Totalevents          int                       `json:"totalEvents"`
	Totalgames           int                       `json:"totalGames"`
	Totalgamesinprogress int                       `json:"totalGamesInProgress"`
	Games                []MLBScheduleResponseGame `json:"games"`
	Events               []interface{}             `json:"events"`
}
