package nfl

type NFLScoreboardResponse struct {
	Leagues []NFLScoreboardLeague `json:"leagues"`
	Events  []NFLScoreboardEvent  `json:"events"`
	Drives  struct {
		Current DriveCurrent `json:"current"`
	} `json:"drives"`
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
