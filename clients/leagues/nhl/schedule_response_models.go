package nhl

import "time"

type NHLScheduleResponse struct {
	NextStartDate          string         `json:"nextStartDate"`
	PreviousStartDate      string         `json:"previousStartDate"`
	GameWeek               []GameWeek     `json:"gameWeek"`
	OddsPartners           []OddsPartners `json:"oddsPartners"`
	PreSeasonStartDate     string         `json:"preSeasonStartDate"`
	RegularSeasonStartDate string         `json:"regularSeasonStartDate"`
	RegularSeasonEndDate   string         `json:"regularSeasonEndDate"`
	PlayoffEndDate         string         `json:"playoffEndDate"`
	NumberOfGames          int            `json:"numberOfGames"`
}
type Venue struct {
	Default string `json:"default"`
}
type TvBroadcasts struct {
	ID          int    `json:"id"`
	Market      string `json:"market"`
	CountryCode string `json:"countryCode"`
	Network     string `json:"network"`
}
type PlaceName struct {
	Default string `json:"default"`
}
type PeriodDescriptor struct {
	Number     int    `json:"number"`
	PeriodType string `json:"periodType"`
}
type Odds struct {
	ProviderID int    `json:"providerId"`
	Value      string `json:"value"`
}

type NHLScheduleTeam struct {
	ID             int       `json:"id,omitempty""`
	PlaceName      PlaceName `json:"placeName,omitempty"`
	Abbrev         string    `json:"abbrev,omitempty"`
	Logo           string    `json:"logo,omitempty"`
	DarkLogo       string    `json:"darkLogo,omitempty"`
	HomeSplitSquad bool      `json:"homeSplitSquad,omitempty"`
	RadioLink      string    `json:"radioLink,omitempty"`
	Odds           []Odds    `json:"odds,omitempty"`
	Score          int       `json:"score,omitempty"`
}
type NHLScheduleResponseGame struct {
	ID                int             `json:"id"`
	Season            int             `json:"season"`
	GameType          int             `json:"gameType"`
	Venue             Venue           `json:"venue"`
	NeutralSite       bool            `json:"neutralSite"`
	StartTimeUTC      time.Time       `json:"startTimeUTC"`
	EasternUTCOffset  string          `json:"easternUTCOffset"`
	VenueUTCOffset    string          `json:"venueUTCOffset"`
	VenueTimezone     string          `json:"venueTimezone"`
	GameState         string          `json:"gameState"`
	GameScheduleState string          `json:"gameScheduleState"`
	TvBroadcasts      []TvBroadcasts  `json:"tvBroadcasts"`
	AwayTeam          NHLScheduleTeam `json:"awayTeam,omitempty"`
	HomeTeam          NHLScheduleTeam `json:"homeTeam,omitempty"`
	GameCenterLink    string          `json:"gameCenterLink"`
	TicketsLink       string          `json:"ticketsLink,omitempty"`
}
type GameWeek struct {
	Date          string                    `json:"date"`
	DayAbbrev     string                    `json:"dayAbbrev"`
	NumberOfGames int                       `json:"numberOfGames"`
	Games         []NHLScheduleResponseGame `json:"games"`
}
type OddsPartners struct {
	PartnerID   int    `json:"partnerId"`
	Country     string `json:"country"`
	Name        string `json:"name"`
	ImageURL    string `json:"imageUrl"`
	SiteURL     string `json:"siteUrl,omitempty"`
	BgColor     string `json:"bgColor"`
	TextColor   string `json:"textColor"`
	AccentColor string `json:"accentColor"`
}
