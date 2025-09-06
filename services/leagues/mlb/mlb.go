package mlb

import (
	"encoding/json"
	"fmt"
	"goalfeed/clients/leagues/mlb"
	"goalfeed/models"
	"goalfeed/utils"
	"strconv"
	"strings"
	"time"
)

type MLBService struct {
	Client mlb.IMLBApiClient
}

const STATUS_UPCOMING = "Preview"
const STATUS_ACTIVE = "Live"
const STATUS_FINAL = "Final"

// const MLB_LEAGUE_ID = 4
var logger = utils.GetLogger()

func (s MLBService) getSchedule() mlb.MLBScheduleResponse {

	//todo implement caching
	//todo support multiple active events
	//todo support some method of determining active events programmatically
	return s.Client.GetMLBSchedule()
}
func (s MLBService) GetLeagueName() string {
	return "MLB"
}

// GetActiveGames Returns active MLBGames
func (s MLBService) GetActiveGames(ret chan []models.Game) {
	schedule := s.getSchedule()
	var activeGames []models.Game

	for _, date := range schedule.Dates {
		for _, game := range date.Games {
			tmpGame := s.gameFromSchedule(game)
			_ = tmpGame
			if gameStatusFromScheduleGame(game) == models.StatusActive {
				activeGames = append(activeGames, s.gameFromSchedule(game))
			}
		}
	}
	ret <- activeGames
}

// GetUpcomingGames Returns upcoming MLBGames
func (s MLBService) GetUpcomingGames(ret chan []models.Game) {
	schedule := s.getSchedule()
	var upcomingGames []models.Game

	for _, date := range schedule.Dates {
		for _, game := range date.Games {
			if gameStatusFromScheduleGame(game) == models.StatusUpcoming {
				upcomingGames = append(upcomingGames, s.gameFromSchedule(game))
			}
		}
	}
	ret <- upcomingGames
}

// GetActiveGames Returns a GameUpdate
func (s MLBService) GetGameUpdate(game models.Game, ret chan models.GameUpdate) {
	if game.CurrentState.ExtTimestamp != "" {
		s.getGameUpdateFromDiffPatch(game, ret)
		// s.getGameUpdateFromScoreboard(game, ret)
	} else {
		s.getGameUpdateFromScoreboard(game, ret)
	}
}
func fudgeTimestamp(extTimestamp string) string {

	pieces := strings.Split(extTimestamp, "_")
	oldTimeInt, _ := strconv.Atoi(pieces[1])
	newTimeInt := oldTimeInt - 10
	_ = pieces
	newTime := fmt.Sprintf("%s_%06d", pieces[0], newTimeInt)
	return newTime

}

func (s MLBService) getGameUpdateFromDiffPatch(game models.Game, ret chan models.GameUpdate) {

	diff, err := s.Client.GetDiffPatch(game.GameCode, fudgeTimestamp(game.CurrentState.ExtTimestamp))
	if err != nil {
		s.getGameUpdateFromScoreboard(game, ret)
		return
	}
	timestampPath := "/metaData/timeStamp"
	homeGoalPath := "/liveData/linescore/teams/home/runs"
	awayGoalPath := "/liveData/linescore/teams/away/runs"
	statusCodePath := "/gameData/status/statusCode"
	var extTimestamp string
	var homeScore int
	var awayScore int
	var statusCode string
	var status models.GameStatus

	for _, set := range diff {
		for _, item := range set.Diff {
			logger.Debug(fmt.Sprintf("Path: %s", item.Path))
			if item.Path == timestampPath {
				json.Unmarshal(item.Value, &extTimestamp)
			} else if item.Path == homeGoalPath {
				logger.Info(fmt.Sprintf("Home score change - %s", game.CurrentState.Home.Team.TeamName))
				json.Unmarshal(item.Value, &homeScore)
			} else if item.Path == awayGoalPath {
				logger.Info(fmt.Sprintf("Away score change - %s", game.CurrentState.Away.Team.TeamName))
				json.Unmarshal(item.Value, &awayScore)
			} else if item.Path == statusCodePath {
				logger.Info("Status Code")
				json.Unmarshal(item.Value, &statusCode)
			}
		}
	}

	if homeScore == 0 {
		homeScore = game.CurrentState.Home.Score
	}
	if awayScore == 0 {
		awayScore = game.CurrentState.Away.Score
	}
	if extTimestamp == "" {
		extTimestamp = game.CurrentState.ExtTimestamp
	}
	if statusCode == "" {
		status = game.CurrentState.Status
	} else {
		status = gameStatusFromStatusCode(statusCode)
	}

	newState := models.GameState{
		Home: models.TeamState{
			Team:  game.CurrentState.Home.Team,
			Score: homeScore,
		},
		Away: models.TeamState{
			Team:  game.CurrentState.Away.Team,
			Score: awayScore,
		},
		Status:       status,
		ExtTimestamp: extTimestamp,
	}

	ret <- models.GameUpdate{
		OldState: game.CurrentState,
		NewState: newState,
	}
}

func (s MLBService) getGameUpdateFromScoreboard(game models.Game, ret chan models.GameUpdate) {
	scoreboard := s.Client.GetMLBScoreBoard(game.GameCode)

	// Extract inning information
	inning := scoreboard.LiveData.Linescore.Currentinning
	isTopInning := scoreboard.LiveData.Linescore.Istopinning

	// Format inning display
	var periodDisplay string
	if inning > 0 {
		if isTopInning {
			periodDisplay = fmt.Sprintf("Top %d", inning)
		} else {
			periodDisplay = fmt.Sprintf("Bot %d", inning)
		}
	}

	// Extract count information
	balls := scoreboard.LiveData.Linescore.Balls
	strikes := scoreboard.LiveData.Linescore.Strikes
	outs := scoreboard.LiveData.Linescore.Outs

	var countDisplay string
	if balls > 0 || strikes > 0 {
		countDisplay = fmt.Sprintf("%d-%d", balls, strikes)
	}
	if outs > 0 {
		if countDisplay != "" {
			countDisplay += fmt.Sprintf(", %d out", outs)
		} else {
			countDisplay = fmt.Sprintf("%d out", outs)
		}
	}

	newState := models.GameState{
		Home: models.TeamState{
			Team:  game.CurrentState.Home.Team,
			Score: scoreboard.LiveData.Linescore.Teams.Home.Runs,
		},
		Away: models.TeamState{
			Team:  game.CurrentState.Away.Team,
			Score: scoreboard.LiveData.Linescore.Teams.Away.Runs,
		},
		Status:        gameStatusFromStatusCode(scoreboard.GameData.Status.StatusCode),
		ExtTimestamp:  scoreboard.MetaData.TimeStamp,
		Period:        inning,
		PeriodType:    "INNING",
		TimeRemaining: countDisplay,
		Clock:         periodDisplay,
		Venue: models.Venue{
			Id:   strconv.Itoa(scoreboard.GameData.Venue.ID),
			Name: scoreboard.GameData.Venue.Name,
		},
		Details: models.EventDetails{
			Inning:      inning,
			Outs:        scoreboard.LiveData.Linescore.Outs,
			BallCount:   scoreboard.LiveData.Linescore.Balls,
			StrikeCount: scoreboard.LiveData.Linescore.Strikes,
			// TODO: Add base runners, pitcher, and batter when available from API
		},
	}
	ret <- models.GameUpdate{
		OldState: game.CurrentState,
		NewState: newState,
	}
}

// MLB team code to ESPN logo URL mapping
func getMLBLogoURL(teamCode string) string {
	return fmt.Sprintf("https://a.espncdn.com/i/teamlogos/mlb/500/%s.png", strings.ToLower(teamCode))
}

func (s MLBService) teamFromScheduleTeam(scheduleTeam mlb.MLBScheduleTeam) models.Team {
	// Get team abbreviation from the team name (MLB API doesn't provide abbreviation in schedule)
	teamCode := s.getTeamCodeFromName(scheduleTeam.Team.Name)

	// Get ESPN logo URL
	logoURL := getMLBLogoURL(teamCode)

	team := models.Team{
		TeamName: scheduleTeam.Team.Name,
		TeamCode: teamCode,
		ExtID:    teamCode,
		LeagueID: models.LeagueIdMLB,
		LogoURL:  logoURL,
	}
	return team
}

func (s MLBService) getTeamCodeFromName(teamName string) string {
	// Map team names to abbreviations
	nameToCode := map[string]string{
		"Arizona Diamondbacks": "ARI", "Atlanta Braves": "ATL", "Baltimore Orioles": "BAL",
		"Boston Red Sox": "BOS", "Chicago Cubs": "CHC", "Chicago White Sox": "CWS",
		"Cincinnati Reds": "CIN", "Cleveland Guardians": "CLE", "Colorado Rockies": "COL",
		"Detroit Tigers": "DET", "Houston Astros": "HOU", "Kansas City Royals": "KC",
		"Los Angeles Angels": "LAA", "Los Angeles Dodgers": "LAD", "Miami Marlins": "MIA",
		"Milwaukee Brewers": "MIL", "Minnesota Twins": "MIN", "New York Mets": "NYM",
		"New York Yankees": "NYY", "Oakland Athletics": "OAK", "Philadelphia Phillies": "PHI",
		"Pittsburgh Pirates": "PIT", "San Diego Padres": "SD", "San Francisco Giants": "SF",
		"Seattle Mariners": "SEA", "St. Louis Cardinals": "STL", "Tampa Bay Rays": "TB",
		"Texas Rangers": "TEX", "Toronto Blue Jays": "TOR", "Washington Nationals": "WSH",
	}

	if code, exists := nameToCode[teamName]; exists {
		return code
	}

	// Fallback: try to extract abbreviation from name
	words := strings.Fields(teamName)
	if len(words) >= 2 {
		// Try last two words for city + team name
		lastTwo := strings.Join(words[len(words)-2:], " ")
		if code, exists := nameToCode[lastTwo]; exists {
			return code
		}
	}

	// Ultimate fallback - return first 3 characters
	if len(teamName) >= 3 {
		return strings.ToUpper(teamName[:3])
	}
	return "UNK"
}
func (s MLBService) gameFromSchedule(scheduleGame mlb.MLBScheduleResponseGame) models.Game {
	// For upcoming games, we don't need detailed scoreboard data - just basic info
	var inning int
	var periodDisplay string
	var countDisplay string
	var venue models.Venue
	var details models.EventDetails

	if scheduleGame.Status.AbstractGameState == STATUS_UPCOMING {
		// For upcoming games, use basic venue info from schedule
		venue = models.Venue{
			Id:   strconv.Itoa(scheduleGame.Venue.ID),
			Name: scheduleGame.Venue.Name,
		}
		details = models.EventDetails{}
	} else {
		// For active games, get detailed scoreboard data
		scoreboard := s.Client.GetMLBScoreBoard(strconv.Itoa(scheduleGame.GamePk))

		// Extract inning information
		inning = scoreboard.LiveData.Linescore.Currentinning
		isTopInning := scoreboard.LiveData.Linescore.Istopinning

		// Format inning display
		if inning > 0 {
			if isTopInning {
				periodDisplay = fmt.Sprintf("Top %d", inning)
			} else {
				periodDisplay = fmt.Sprintf("Bot %d", inning)
			}
		}

		// Extract count information
		balls := scoreboard.LiveData.Linescore.Balls
		strikes := scoreboard.LiveData.Linescore.Strikes
		outs := scoreboard.LiveData.Linescore.Outs

		if balls > 0 || strikes > 0 {
			countDisplay = fmt.Sprintf("%d-%d", balls, strikes)
		}
		if outs > 0 {
			if countDisplay != "" {
				countDisplay += fmt.Sprintf(", %d out", outs)
			} else {
				countDisplay = fmt.Sprintf("%d out", outs)
			}
		}

		venue = models.Venue{
			Id:   strconv.Itoa(scoreboard.GameData.Venue.ID),
			Name: scoreboard.GameData.Venue.Name,
		}

		details = models.EventDetails{
			Inning:      inning,
			Outs:        scoreboard.LiveData.Linescore.Outs,
			BallCount:   scoreboard.LiveData.Linescore.Balls,
			StrikeCount: scoreboard.LiveData.Linescore.Strikes,
			// TODO: Add base runners, pitcher, and batter when available from API
		}
	}

	// Convert UTC time to local time for display
	localTime := scheduleGame.Gamedate.Local()

	// Format game time for upcoming games
	var gameTimeDisplay string
	if scheduleGame.Status.AbstractGameState == STATUS_UPCOMING {
		gameTimeDisplay = localTime.Format("Mon 3:04 PM")
	} else {
		gameTimeDisplay = periodDisplay
	}

	return models.Game{
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team:  s.teamFromScheduleTeam(scheduleGame.Teams.Home),
				Score: scheduleGame.Teams.Home.Score,
			},
			Away: models.TeamState{
				Team:  s.teamFromScheduleTeam(scheduleGame.Teams.Away),
				Score: scheduleGame.Teams.Away.Score,
			},
			Status:        gameStatusFromScheduleGame(scheduleGame),
			FetchedAt:     time.Now(),
			Period:        inning,
			PeriodType:    "INNING",
			TimeRemaining: countDisplay,
			Clock:         gameTimeDisplay,
			Venue:         venue,
			Details:       details,
		},
		GameCode: strconv.Itoa(scheduleGame.GamePk),
		LeagueId: models.LeagueIdMLB,
		GameDetails: models.GameDetails{
			GameId:     strconv.Itoa(scheduleGame.GamePk),
			Season:     scheduleGame.Season,
			SeasonType: scheduleGame.Gametype,
			GameDate:   localTime,
			GameTime:   localTime.Format("3:04 PM"),
			Timezone:   "Local",
		},
	}
}
func gameStatusFromScheduleGame(scheduleGame mlb.MLBScheduleResponseGame) models.GameStatus {
	switch scheduleGame.Status.AbstractGameState {
	case STATUS_FINAL:
		return models.StatusEnded
	case STATUS_UPCOMING:
		return models.StatusUpcoming
	case STATUS_ACTIVE:
		return models.StatusActive
	default:
		return models.StatusActive
	}
}
func gameStatusFromStatusCode(statusCode string) models.GameStatus {
	switch statusCode {
	case "7":
		return models.StatusEnded
	default:
		return models.StatusActive
	}
}
func (s MLBService) GetEvents(update models.GameUpdate, ret chan []models.Event) {
	events := append(
		s.getGoalEvents(update.OldState.Home, update.NewState.Home, update.OldState.Away.Team),
		s.getGoalEvents(update.OldState.Away, update.NewState.Away, update.OldState.Home.Team)...,
	)
	ret <- events
}
func (s MLBService) getGoalEvents(oldState models.TeamState, newState models.TeamState, opponent models.Team) []models.Event {
	events := []models.Event{}
	diff := newState.Score - oldState.Score
	if diff <= 0 {
		return events
	}
	team := newState.Team

	for i := 0; i < diff; i++ {
		events = append(events, models.Event{
			TeamCode:     team.TeamCode,
			TeamName:     team.TeamName,
			TeamHash:     team.GetTeamHash(),
			LeagueId:     models.LeagueIdMLB,
			LeagueName:   s.GetLeagueName(),
			OpponentCode: opponent.TeamCode,
			OpponentName: opponent.TeamName,
			OpponentHash: opponent.GetTeamHash(),
		})
	}
	return events
}
