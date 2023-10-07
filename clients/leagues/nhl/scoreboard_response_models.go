package nhl

import (
	"time"
)

type ShootoutTeamInfo struct {
	Scores   int `json:"scores"`
	Attempts int `json:"attempts"`
}
type LinescoreTeam struct {
	Team struct {
		ID           int    `json:"id"`
		Name         string `json:"name"`
		Link         string `json:"link"`
		Abbreviation string `json:"abbreviation"`
		TriCode      string `json:"triCode"`
	} `json:"team"`
	Goals        int  `json:"goals"`
	ShotsOnGoal  int  `json:"shotsOnGoal"`
	GoaliePulled bool `json:"goaliePulled"`
	NumSkaters   int  `json:"numSkaters"`
	PowerPlay    bool `json:"powerPlay"`
}

type GameDataTeam struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Link  string `json:"link"`
	Venue struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Link     string `json:"link"`
		City     string `json:"city"`
		TimeZone struct {
			ID     string `json:"id"`
			Offset int    `json:"offset"`
			Tz     string `json:"tz"`
		} `json:"timeZone"`
	} `json:"venue"`
	Abbreviation    string `json:"abbreviation"`
	TriCode         string `json:"triCode"`
	TeamName        string `json:"teamName"`
	LocationName    string `json:"locationName"`
	FirstYearOfPlay string `json:"firstYearOfPlay"`
	Division        struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Link string `json:"link"`
	} `json:"division"`
	Conference struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Link string `json:"link"`
	} `json:"conference"`
	Franchise struct {
		FranchiseID int    `json:"franchiseId"`
		TeamName    string `json:"teamName"`
		Link        string `json:"link"`
	} `json:"franchise"`
	ShortName       string `json:"shortName"`
	OfficialSiteURL string `json:"officialSiteUrl"`
	FranchiseID     int    `json:"franchiseId"`
	Active          bool   `json:"active"`
}
type NHLScoreboardResponse struct {
	GamePk    int    `json:"gamePk"`
	Link      string `json:"link"`
	MetaData  struct {
		Wait      int    `json:"wait"`
		TimeStamp string `json:"timeStamp"`
	} `json:"metaData"`
	GameData struct {
		Game struct {
			Pk     int    `json:"pk"`
			Season string `json:"season"`
			Type   string `json:"type"`
		} `json:"game"`
		Datetime struct {
			DateTime time.Time `json:"dateTime"`
		} `json:"datetime"`
		Status struct {
			AbstractGameState string `json:"abstractGameState"`
			CodedGameState    string `json:"codedGameState"`
			DetailedState     string `json:"detailedState"`
			StatusCode        string `json:"statusCode"`
			StartTimeTBD      bool   `json:"startTimeTBD"`
		} `json:"status"`
		Teams struct {
			Away GameDataTeam `json:"away"`
			Home GameDataTeam `json:"home"`
		} `json:"teams"`
	} `json:"gameData"`
	LiveData struct {
		Linescore struct {
			CurrentPeriod int           `json:"currentPeriod"`
			Periods       []interface{} `json:"periods"`
			ShootoutInfo  struct {
				Away ShootoutTeamInfo `json:"away"`
				Home ShootoutTeamInfo `json:"home"`
			} `json:"shootoutInfo"`
			Teams struct {
				Home LinescoreTeam `json:"home"`
				Away LinescoreTeam `json:"away"`
			} `json:"teams"`
			PowerPlayStrength string `json:"powerPlayStrength"`
			HasShootout       bool   `json:"hasShootout"`
			IntermissionInfo  struct {
				IntermissionTimeRemaining int  `json:"intermissionTimeRemaining"`
				IntermissionTimeElapsed   int  `json:"intermissionTimeElapsed"`
				InIntermission            bool `json:"inIntermission"`
			} `json:"intermissionInfo"`
		} `json:"linescore"`
	} `json:"liveData"`
}
