package nhl

type NHLTeamResponse struct {
	Teams []NHLTeamsResponseTeam `json:"teams"`
}
type NHLTeamsResponseTeam struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Link            string `json:"link"`
	Abbreviation    string `json:"abbreviation"`
	TeamName        string `json:"teamName"`
	LocationName    string `json:"locationName"`
	ShortName       string `json:"shortName"`
	OfficialSiteURL string `json:"officialSiteUrl"`
	FranchiseID     int    `json:"franchiseId"`
	Active          bool   `json:"active"`
}
