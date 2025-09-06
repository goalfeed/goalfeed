package cfl

// CFL Schedule Response Models
type CFLScheduleResponse struct {
	Rounds []CFLRound `json:"rounds"`
}

type CFLRound struct {
	ID          int           `json:"id"`
	Status      string        `json:"status"`
	Name        string        `json:"name"`
	Type        string        `json:"type"`
	Number      int           `json:"number"`
	StartDate   string        `json:"startDate"`
	EndDate     string        `json:"endDate"`
	Tournaments []CFLGame     `json:"tournaments"`
}

type CFLGame struct {
	ID           int           `json:"id"`
	Date         string        `json:"date"`
	Status       string        `json:"status"`
	HomeSquad    CFLTeam       `json:"homeSquad"`
	AwaySquad    CFLTeam       `json:"awaySquad"`
	ActivePeriod interface{}   `json:"activePeriod"`
	Timeouts     CFLTimeouts   `json:"timeouts"`
	Possession   string        `json:"possession"`
	CFLID        int           `json:"cflId"`
	Clock        string        `json:"clock"`
	Winner       interface{}   `json:"winner"`
	IsHidden     bool          `json:"isHidden"`
	Markets      interface{}   `json:"markets"`
	MarketsBCLC  []interface{} `json:"marketsBCLC"`
}

type CFLTeam struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	ShortName string `json:"shortName"`
	Score    int    `json:"score"`
}

type CFLTimeouts struct {
	Away int `json:"away"`
	Home int `json:"home"`
}

// CFL Live Game Response Models (from BetGenius API)
type CFLLiveGameResponse struct {
	Data CFLLiveGameData `json:"data"`
	Sport string `json:"sport"`
	SportID int `json:"sportId"`
	CompetitionID int `json:"competitionId"`
	AvailableTabs CFLAvailableTabs `json:"availableTabs"`
}

type CFLLiveGameData struct {
	BetGeniusFixtureID string `json:"betGeniusFixtureId"`
	PreferredSourceIDs CFLPreferredSourceIDs `json:"preferredSourceIds"`
	ScoreboardInfo CFLScoreboardInfo `json:"scoreboardInfo"`
	MatchInfo CFLMatchInfo `json:"matchInfo"`
	Court CFLCourt `json:"court"`
	Scheduler bool `json:"scheduler"`
}

type CFLPreferredSourceIDs struct {
	MatchActionsSourceID string `json:"matchActionsSourceId"`
	PlayerStatsSourceID string `json:"playerStatsSourceId"`
	TeamStatsSourceID string `json:"teamStatsSourceId"`
}

type CFLScoreboardInfo struct {
	MatchStatus string `json:"matchStatus"`
	CurrentPhase string `json:"currentPhase"`
	AwayScore int `json:"awayScore"`
	HomeScore int `json:"homeScore"`
	AwayTimeoutsLeft int `json:"awayTimeoutsLeft"`
	HomeTimeoutsLeft int `json:"homeTimeoutsLeft"`
	TotalTimeouts int `json:"totalTimeouts"`
	ScoreByPhases CFLScoreByPhases `json:"scoreByPhases"`
	TimeRemainingInPhase string `json:"timeRemainingInPhase"`
	Possession string `json:"possession"`
	Down interface{} `json:"down"`
	YardsToGo interface{} `json:"yardsToGo"`
	TotalPhases int `json:"totalPhases"`
	PhaseQualifier string `json:"phaseQualifier"`
	ClockUnreliable bool `json:"clockUnreliable"`
}

type CFLScoreByPhases struct {
	AwayScore CFLPhaseScore `json:"awayScore"`
	HomeScore CFLPhaseScore `json:"homeScore"`
}

type CFLPhaseScore struct {
	Quarter1 int `json:"quarter1"`
}

type CFLMatchInfo struct {
	RoundID string `json:"roundId"`
	RoundName string `json:"roundName"`
	ScheduledStartTime string `json:"scheduledStartTime"`
	VenueName string `json:"venueName"`
	SeasonID string `json:"seasonId"`
	SeasonName string `json:"seasonName"`
	HomeTeam CFLDetailedTeam `json:"homeTeam"`
	AwayTeam CFLDetailedTeam `json:"awayTeam"`
	PlayedPhases []string `json:"playedPhases"`
}

type CFLDetailedTeam struct {
	FullName string `json:"fullName"`
	CompetitorID string `json:"competitorId"`
	Details CFLTeamDetails `json:"details"`
}

type CFLTeamDetails struct {
	Key string `json:"key"`
	Brand CFLBrand `json:"brand"`
	PrimaryColor string `json:"primaryColor"`
	SecondaryColor string `json:"secondaryColor"`
	FirstName string `json:"firstName"`
	ShortName string `json:"shortName"`
	SecondName string `json:"secondName"`
	Abbreviation string `json:"abbreviation"`
	OfficialName string `json:"officialName"`
}

type CFLBrand struct {
	Logo string `json:"logo"`
	Theme CFLTheme `json:"theme"`
}

type CFLTheme struct {
	Dark CFLThemeColors `json:"dark"`
	Light CFLThemeColors `json:"light"`
}

type CFLThemeColors struct {
	Logo CFLThemeLogo `json:"logo"`
	PrimaryColor string `json:"primaryColor"`
	SecondaryColor string `json:"secondaryColor"`
}

type CFLThemeLogo struct {
	SVG string `json:"svg"`
}

type CFLCourt struct {
	MatchActions []interface{} `json:"matchActions"`
}

type CFLAvailableTabs struct {
	Court bool `json:"court"`
	TeamStats bool `json:"teamStats"`
	PlayerStats bool `json:"playerStats"`
	Lineups bool `json:"lineups"`
	PlayByPlay bool `json:"playByPlay"`
}
