package nfl

type NFLTeamResponse struct {
	Teams []NFLTeam `json:"teams"`
}

type NFLTeam struct {
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
		ID       string `json:"id"`
		FullName string `json:"fullName"`
		Address  struct {
			City  string `json:"city"`
			State string `json:"state"`
		} `json:"address"`
		Grass  bool `json:"grass"`
		Indoor bool `json:"indoor"`
	} `json:"venue"`
	Links []struct {
		Rel  []string `json:"rel"`
		Href string   `json:"href"`
		Text string   `json:"text"`
	} `json:"links"`
	Logo string `json:"logo"`
}

