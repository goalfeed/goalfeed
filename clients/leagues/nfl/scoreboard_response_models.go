package nfl

// NFLScoreboardResponse models the JSON returned by GetNFLScoreBoard.
//
// IMPORTANT: this type is shared by two genuinely different ESPN response
// shapes:
//
//   - The "/summary?event=<id>" endpoint, which is what GetNFLScoreBoard
//     actually calls in production. It has NO top-level "events" array at
//     all; the live score/status/team data lives under
//     "header.competitions[0]" (see Header/GameInfo below). Verified against
//     a live event: top-level keys are againstTheSpread, boxscore,
//     boxscore, drives, format, gameInfo, header, injuries, leaders, meta,
//     news, odds, pickcenter, scoringPlays, standings, videos,
//     wallclockAvailable, winprobability -- no "events" key.
//   - The "/scoreboard" endpoint (used by GetNFLSchedule /
//     GetNFLScheduleByDate), which DOES have a top-level "events" array.
//     That endpoint doesn't return an NFLScoreboardResponse (it returns
//     NFLScheduleResponse), but the mocks/tests in this package model
//     GetNFLScoreBoard's return value using the Events shape for
//     historical reasons, so Events is kept here and used as a fallback.
//
// Parsing code (see services/leagues/nfl.extractSummarySnapshot) prefers
// Header when present -- the real shape returned by the live HTTP call --
// and falls back to Events only for compatibility with existing
// mocks/tests. Events will always be empty for real traffic.
type NFLScoreboardResponse struct {
	Leagues []NFLScoreboardLeague `json:"leagues"`
	Events  []NFLScoreboardEvent  `json:"events"`
	Drives  struct {
		Current DriveCurrent `json:"current"`
	} `json:"drives"`
	Header   NFLSummaryHeader   `json:"header"`
	GameInfo NFLSummaryGameInfo `json:"gameInfo"`
}

// NFLSummaryHeader models "header" from the real "/summary?event=" payload.
//
// NOTE: unlike NFLScheduleEvent.Week / NFLScoreboardEvent.Week (which are
// both {"number": N} objects on the "/scoreboard" endpoint), the summary
// endpoint's "header.week" is a bare integer -- verified against a live
// event ("week": 3, not {"number": 3}). Modeling it as int directly avoids
// the same silent-whole-payload-discard failure mode as the CFL bug.
type NFLSummaryHeader struct {
	ID     string `json:"id"`
	Season struct {
		Year int `json:"year"`
	} `json:"season"`
	Week         int                     `json:"week"`
	Competitions []NFLSummaryCompetition `json:"competitions"`
}

// NFLSummaryCompetition models "header.competitions[]". Competitors reuses
// NFLScoreboardCompetitor -- the real payload's competitor/team fields
// (id, homeAway, winner, score, team.abbreviation, team.displayName, ...)
// line up with that type's field names and types; extra real-world fields
// (e.g. team.logos[] instead of a single team.logo) are simply additive and
// don't cause unmarshal errors.
type NFLSummaryCompetition struct {
	ID          string                    `json:"id"`
	Date        string                    `json:"date"`
	Competitors []NFLScoreboardCompetitor `json:"competitors"`
	Status      NFLSummaryStatus          `json:"status"`
}

// NFLSummaryStatus models "header.competitions[].status" from the summary
// endpoint. Period/DisplayClock are present while a game is live; a
// finished game's status only carries Type. Type.State/.Detail are the
// authoritative status signal ("pre"/"in"/"post", e.g. detail
// "4:57 - 3rd Quarter" while live, "Final" once complete).
type NFLSummaryStatus struct {
	Period       int    `json:"period"`
	DisplayClock string `json:"displayClock"`
	Type         struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		State       string `json:"state"`
		Completed   bool   `json:"completed"`
		Description string `json:"description"`
		Detail      string `json:"detail"`
		ShortDetail string `json:"shortDetail"`
	} `json:"type"`
}

// NFLSummaryGameInfo models "gameInfo" from the summary endpoint, which is
// where venue info actually lives (NOT competition.venue -- that field is
// absent from the real summary payload).
type NFLSummaryGameInfo struct {
	Venue struct {
		FullName string `json:"fullName"`
	} `json:"venue"`
}

type NFLScoreboardLeague struct {
	ID           string `json:"id"`
	UID          string `json:"uid"`
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
	Slug         string `json:"slug"`
	Season       struct {
		Year int `json:"year"`
	} `json:"season"`
}

type NFLScoreboardEvent struct {
	ID        string `json:"id"`
	UID       string `json:"uid"`
	Date      string `json:"date"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	Season    struct {
		Year int `json:"year"`
	} `json:"season"`
	Week struct {
		Number int `json:"number"`
	} `json:"week"`
	Competitions []NFLScoreboardCompetition `json:"competitions"`
	Status       struct {
		Clock        float64 `json:"clock"`
		DisplayClock string  `json:"displayClock"`
		Period       int     `json:"period"`
		Type         struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			State       string `json:"state"`
			Completed   bool   `json:"completed"`
			Description string `json:"description"`
			Detail      string `json:"detail"`
			ShortDetail string `json:"shortDetail"`
		} `json:"type"`
	} `json:"status"`
}

type NFLScoreboardCompetition struct {
	ID         string `json:"id"`
	UID        string `json:"uid"`
	Date       string `json:"date"`
	Attendance int    `json:"attendance"`
	Type       struct {
		ID           string `json:"id"`
		Abbreviation string `json:"abbreviation"`
	} `json:"type"`
	TimeValid             bool `json:"timeValid"`
	NeutralSite           bool `json:"neutralSite"`
	ConferenceCompetition bool `json:"conferenceCompetition"`
	PlayByPlayAvailable   bool `json:"playByPlayAvailable"`
	Recent                bool `json:"recent"`
	Venue                 struct {
		ID       string `json:"id"`
		FullName string `json:"fullName"`
		Address  struct {
			City  string `json:"city"`
			State string `json:"state"`
		} `json:"address"`
		Grass  bool `json:"grass"`
		Indoor bool `json:"indoor"`
	} `json:"venue"`
	Competitors []NFLScoreboardCompetitor `json:"competitors"`
	Notes       []struct {
		Headline string `json:"headline"`
	} `json:"notes"`
}

type NFLScoreboardCompetitor struct {
	ID       string `json:"id"`
	UID      string `json:"uid"`
	Type     string `json:"type"`
	Order    int    `json:"order"`
	HomeAway string `json:"homeAway"`
	Winner   bool   `json:"winner"`
	Team     struct {
		ID               string `json:"id"`
		UID              string `json:"uid"`
		Location         string `json:"location"`
		Name             string `json:"name"`
		Abbreviation     string `json:"abbreviation"`
		DisplayName      string `json:"displayName"`
		ShortDisplayName string `json:"shortDisplayName"`
		Color            string `json:"color"`
		AlternateColor   string `json:"alternateColor"`
		IsActive         bool   `json:"isActive"`
		Venue            struct {
			ID string `json:"id"`
		} `json:"venue"`
		Links []struct {
			Rel  []string `json:"rel"`
			Href string   `json:"href"`
			Text string   `json:"text"`
		} `json:"links"`
		// Logo is populated by the "/scoreboard" endpoint's team object.
		// The real summary endpoint's team object instead carries a
		// "logos" array of variants (no single "logo" string) -- that
		// array isn't modeled here since logo hydration isn't part of the
		// score/status/clock bug this type was extended to fix, and the
		// extra field is simply ignored (not an unmarshal error) rather
		// than populated.
		Logo string `json:"logo"`
	} `json:"team"`
	Score      string `json:"score"`
	Linescores []struct {
		Value float64 `json:"value"`
	} `json:"linescores"`
	Statistics []struct {
		Label string `json:"label"`
		Stats []struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"stats"`
	} `json:"statistics"`
	Records []struct {
		Name         string `json:"name"`
		Abbreviation string `json:"abbreviation"`
		Type         string `json:"type"`
		Summary      string `json:"summary"`
	} `json:"records"`
}

// Drives and current drive start info

type DriveCurrent struct {
	ID          string     `json:"id"`
	Description string     `json:"description"`
	Team        DriveTeam  `json:"team"`
	Start       DriveStart `json:"start"`
}

type DriveTeam struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
}

type DriveStart struct {
	Down                  int    `json:"down"`
	Distance              int    `json:"distance"`
	YardLine              int    `json:"yardLine"`
	YardsToEndzone        int    `json:"yardsToEndzone"`
	DownDistanceText      string `json:"downDistanceText"`
	ShortDownDistanceText string `json:"shortDownDistanceText"`
	PossessionText        string `json:"possessionText"`
	Team                  struct {
		ID string `json:"id"`
	} `json:"team"`
}
