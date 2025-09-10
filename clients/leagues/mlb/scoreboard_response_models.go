package mlb

import (
	"time"
)

type MLBScoreboardResponse struct {
	Copyright string   `json:"copyright"`
	Gamepk    int      `json:"gamePk"`
	Link      string   `json:"link"`
	MetaData  Metadata `json:"metaData"`
	GameData  GameData `json:"gameData"`
	LiveData  LiveData `json:"liveData"`
}
type Metadata struct {
	Wait          int      `json:"wait"`
	TimeStamp     string   `json:"timeStamp"`
	Gameevents    []string `json:"gameEvents"`
	Logicalevents []string `json:"logicalEvents"`
}
type Game struct {
	Pk              int    `json:"pk"`
	Type            string `json:"type"`
	Doubleheader    string `json:"doubleHeader"`
	ID              string `json:"id"`
	Gamedaytype     string `json:"gamedayType"`
	Tiebreaker      string `json:"tiebreaker"`
	Gamenumber      int    `json:"gameNumber"`
	Calendareventid string `json:"calendarEventID"`
	Season          string `json:"season"`
	Seasondisplay   string `json:"seasonDisplay"`
}
type Datetime struct {
	Datetime     time.Time `json:"dateTime"`
	Originaldate string    `json:"originalDate"`
	Daynight     string    `json:"dayNight"`
	Time         string    `json:"time"`
	Ampm         string    `json:"ampm"`
}
type League struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Link string `json:"link"`
}
type Division struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Link string `json:"link"`
}
type Sport struct {
	ID   int    `json:"id"`
	Link string `json:"link"`
	Name string `json:"name"`
}
type Leaguerecord struct {
	Wins   int    `json:"wins"`
	Losses int    `json:"losses"`
	Pct    string `json:"pct"`
}
type MLBGameResponseTeam struct {
	ID              int      `json:"id"`
	Name            string   `json:"name"`
	Link            string   `json:"link"`
	Runs            int      `json:"runs"`
	Season          int      `json:"season"`
	Venue           Venue    `json:"venue"`
	Teamcode        string   `json:"teamCode"`
	Filecode        string   `json:"fileCode"`
	Abbreviation    string   `json:"abbreviation"`
	Teamname        string   `json:"teamName"`
	Locationname    string   `json:"locationName"`
	Firstyearofplay string   `json:"firstYearOfPlay"`
	League          League   `json:"league"`
	Division        Division `json:"division"`
	Sport           Sport    `json:"sport"`
	Shortname       string   `json:"shortName"`
	Allstarstatus   string   `json:"allStarStatus"`
	Active          bool     `json:"active"`
	Team            Team     `json:"team"`
}
type MLBScoreboardResponseTeams struct {
	Away MLBGameResponseTeam `json:"away"`
	Home MLBGameResponseTeam `json:"home"`
}

type GameData struct {
	Game     Game                       `json:"game"`
	Datetime Datetime                   `json:"datetime"`
	Status   Status                     `json:"status"`
	Teams    MLBScoreboardResponseTeams `json:"teams"`
	Venue    Venue                      `json:"venue"`
	Alerts   []interface{}              `json:"alerts"`
}

//	type Result struct {
//		Type        string `json:"type"`
//		Event       string `json:"event"`
//		Eventtype   string `json:"eventType"`
//		Description string `json:"description"`
//		Rbi         int    `json:"rbi"`
//		Awayscore   int    `json:"awayScore"`
//		Homescore   int    `json:"homeScore"`
//	}
type Details struct {
	Description   string `json:"description"`
	Event         string `json:"event"`
	Eventtype     string `json:"eventType"`
	Awayscore     int    `json:"awayScore"`
	Homescore     int    `json:"homeScore"`
	Isscoringplay bool   `json:"isScoringPlay"`
	Hasreview     bool   `json:"hasReview"`
}
type Result struct {
	Type      string `json:"type"`
	Rbi       int    `json:"rbi"`
	AwayScore int    `json:"awayScore"`
	HomeScore int    `json:"homeScore"`
}
type About struct {
	Atbatindex       int       `json:"atBatIndex"`
	Halfinning       string    `json:"halfInning"`
	Istopinning      bool      `json:"isTopInning"`
	Inning           int       `json:"inning"`
	Starttime        time.Time `json:"startTime"`
	Endtime          time.Time `json:"endTime"`
	Iscomplete       bool      `json:"isComplete"`
	Isscoringplay    bool      `json:"isScoringPlay"`
	Hasout           bool      `json:"hasOut"`
	Captivatingindex int       `json:"captivatingIndex"`
}
type Team struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Link          string `json:"link"`
	Allstarstatus string `json:"allStarStatus"`
}
type Innings struct {
	Num        int                 `json:"num"`
	Ordinalnum string              `json:"ordinalNum"`
	Home       MLBGameResponseTeam `json:"home,omitempty"`
	Away       MLBGameResponseTeam `json:"away,omitempty"`
}
type Linescore struct {
	Currentinning        int                        `json:"currentInning"`
	Currentinningordinal string                     `json:"currentInningOrdinal"`
	Inningstate          string                     `json:"inningState"`
	Inninghalf           string                     `json:"inningHalf"`
	Istopinning          bool                       `json:"isTopInning"`
	Scheduledinnings     int                        `json:"scheduledInnings"`
	Innings              []Innings                  `json:"innings"`
	Teams                MLBScoreboardResponseTeams `json:"teams"`
	Balls                int                        `json:"balls"`
	Strikes              int                        `json:"strikes"`
	Outs                 int                        `json:"outs"`
}
type Info struct {
	Label string `json:"label"`
	Value string `json:"value,omitempty"`
}
type Boxscore struct {
	Teams         BoxscoreTeams `json:"teams"`
	Info          []Info        `json:"info"`
	Pitchingnotes []interface{} `json:"pitchingNotes"`
}
type LiveData struct {
	Linescore Linescore `json:"linescore"`
	Boxscore  Boxscore  `json:"boxscore"`
}

// Player information structures
type Player struct {
	Person       Person      `json:"person"`
	JerseyNumber string      `json:"jerseyNumber"`
	Position     Position    `json:"position"`
	Status       Status      `json:"status"`
	ParentTeamID int         `json:"parentTeamId"`
	Stats        PlayerStats `json:"stats"`
	SeasonStats  PlayerStats `json:"seasonStats"`
	GameStatus   GameStatus  `json:"gameStatus"`
}

type Person struct {
	ID       int    `json:"id"`
	FullName string `json:"fullName"`
	Link     string `json:"link"`
}

type Position struct {
	Code         string `json:"code"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Abbreviation string `json:"abbreviation"`
}

type PlayerStats struct {
	Batting  map[string]interface{} `json:"batting"`
	Pitching map[string]interface{} `json:"pitching"`
	Fielding map[string]interface{} `json:"fielding"`
}

type GameStatus struct {
	IsCurrentBatter  bool `json:"isCurrentBatter"`
	IsCurrentPitcher bool `json:"isCurrentPitcher"`
	IsOnBench        bool `json:"isOnBench"`
	IsSubstitute     bool `json:"isSubstitute"`
}

type TeamBoxscore struct {
	Team         Team              `json:"team"`
	Players      map[string]Player `json:"players"`
	Batters      []int             `json:"batters"`
	Pitchers     []int             `json:"pitchers"`
	Bench        []int             `json:"bench"`
	Bullpen      []int             `json:"bullpen"`
	BattingOrder []int             `json:"battingOrder"`
	Info         []Info            `json:"info"`
	Note         []interface{}     `json:"note"`
}

type BoxscoreTeams struct {
	Away TeamBoxscore `json:"away"`
	Home TeamBoxscore `json:"home"`
}
