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

func (s MLBService) getScheduleByDate(date string) mlb.MLBScheduleResponse {
	return s.Client.GetMLBScheduleByDate(date)
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
			status := gameStatusFromScheduleGame(game)
			if status == models.StatusActive || status == models.StatusDelayed {
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

// GetGamesByDate Returns all MLB games for a specific date
func (s MLBService) GetGamesByDate(date string, ret chan []models.Game) {
	schedule := s.getScheduleByDate(date)
	var games []models.Game

	for _, dateGroup := range schedule.Dates {
		for _, game := range dateGroup.Games {
			// Include all games for the specified date
			games = append(games, s.gameFromSchedule(game))
		}
	}
	ret <- games
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
		// Preserve existing detailed game state when using diff patch updates
		Period:        game.CurrentState.Period,
		PeriodType:    game.CurrentState.PeriodType,
		TimeRemaining: game.CurrentState.TimeRemaining,
		Clock:         game.CurrentState.Clock,
		Venue:         game.CurrentState.Venue,
		Details:       game.CurrentState.Details,
		Statistics:    game.CurrentState.Statistics,
	}

	ret <- models.GameUpdate{
		OldState: game.CurrentState,
		NewState: newState,
	}
}

func (s MLBService) getGameUpdateFromScoreboard(game models.Game, ret chan models.GameUpdate) {
	logger.Info(fmt.Sprintf("Getting scoreboard update for game %s", game.GameCode))
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
			Statistics: models.TeamStats{
				Hits:   scoreboard.LiveData.Linescore.Teams.Home.Hits,
				Errors: scoreboard.LiveData.Linescore.Teams.Home.Errors,
			},
		},
		Away: models.TeamState{
			Team:  game.CurrentState.Away.Team,
			Score: scoreboard.LiveData.Linescore.Teams.Away.Runs,
			Statistics: models.TeamStats{
				Hits:   scoreboard.LiveData.Linescore.Teams.Away.Hits,
				Errors: scoreboard.LiveData.Linescore.Teams.Away.Errors,
			},
		},
		Status: func() models.GameStatus {
			status := gameStatusFromStatusCode(scoreboard.GameData.Status.StatusCode)
			logger.Info(fmt.Sprintf("Scoreboard status for game %s: %s -> %d",
				game.GameCode, scoreboard.GameData.Status.StatusCode, status))
			return status
		}(),
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
			BaseRunners: func() models.BaseRunners {
				off := scoreboard.LiveData.Linescore.Offense
				convert := func(or *mlb.OffenseRunner) *models.Player {
					if or == nil {
						return nil
					}
					return &models.Player{ // Position and jersey unknown here
						Id:       strconv.Itoa(or.ID),
						Name:     or.FullName,
						Number:   0,
						Position: "",
						Team:     models.Team{TeamCode: "", TeamName: "", LeagueID: models.LeagueIdMLB},
					}
				}
				return models.BaseRunners{
					First:  convert(off.First),
					Second: convert(off.Second),
					Third:  convert(off.Third),
				}
			}(),
			// Extract current pitcher and batter information
			Pitcher: s.extractCurrentPitcher(scoreboard.LiveData.Boxscore.Teams),
			Batter:  s.extractCurrentBatter(scoreboard.LiveData.Boxscore.Teams),
		},
	}

	// Populate total statistics for convenience (used by frontend GameCard)
	newState.Statistics = models.TeamStats{
		Hits:   newState.Home.Statistics.Hits + newState.Away.Statistics.Hits,
		Errors: newState.Home.Statistics.Errors + newState.Away.Statistics.Errors,
	}

	// Override to active when inning/count indicate gameplay but status mapping did not
	if newState.Status == models.StatusUpcoming {
		if inning > 0 || periodDisplay != "" || outs > 0 {
			newState.Status = models.StatusActive
		}
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
		logger.Info(fmt.Sprintf("Game %d processed as UPCOMING", scheduleGame.GamePk))
		venue = models.Venue{
			Id:   strconv.Itoa(scheduleGame.Venue.ID),
			Name: scheduleGame.Venue.Name,
		}
		details = models.EventDetails{}
	} else {
		// For active games, get detailed scoreboard data
		logger.Info(fmt.Sprintf("Game %d processed as ACTIVE - AbstractGameState: %s",
			scheduleGame.GamePk, scheduleGame.Status.AbstractGameState))
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
			BaseRunners: func() models.BaseRunners {
				off := scoreboard.LiveData.Linescore.Offense
				convert := func(or *mlb.OffenseRunner) *models.Player {
					if or == nil {
						return nil
					}
					return &models.Player{
						Id:       strconv.Itoa(or.ID),
						Name:     or.FullName,
						Number:   0,
						Position: "",
						Team:     models.Team{TeamCode: "", TeamName: "", LeagueID: models.LeagueIdMLB},
					}
				}
				return models.BaseRunners{First: convert(off.First), Second: convert(off.Second), Third: convert(off.Third)}
			}(),
			// Extract current pitcher and batter information
			Pitcher: s.extractCurrentPitcher(scoreboard.LiveData.Boxscore.Teams),
			Batter:  s.extractCurrentBatter(scoreboard.LiveData.Boxscore.Teams),
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

	// Build initial team statistics from linescore if available
	homeStats := models.TeamStats{}
	awayStats := models.TeamStats{}
	if scheduleGame.Status.AbstractGameState != STATUS_UPCOMING {
		// Re-fetch scoreboard for stats context in this scope
		sb := s.Client.GetMLBScoreBoard(strconv.Itoa(scheduleGame.GamePk))
		homeStats.Hits = sb.LiveData.Linescore.Teams.Home.Hits
		homeStats.Errors = sb.LiveData.Linescore.Teams.Home.Errors
		awayStats.Hits = sb.LiveData.Linescore.Teams.Away.Hits
		awayStats.Errors = sb.LiveData.Linescore.Teams.Away.Errors
	}

	return models.Game{
		CurrentState: models.GameState{
			Home: models.TeamState{
				Team:       s.teamFromScheduleTeam(scheduleGame.Teams.Home),
				Score:      scheduleGame.Teams.Home.Score,
				Statistics: homeStats,
			},
			Away: models.TeamState{
				Team:       s.teamFromScheduleTeam(scheduleGame.Teams.Away),
				Score:      scheduleGame.Teams.Away.Score,
				Statistics: awayStats,
			},
			Status: func() models.GameStatus {
				status := gameStatusFromScheduleGame(scheduleGame)
				logger.Info(fmt.Sprintf("Game %d final status: %d (from schedule: AbstractGameState=%s, DetailedState=%s, StatusCode=%s)",
					scheduleGame.GamePk, status, scheduleGame.Status.AbstractGameState,
					scheduleGame.Status.DetailedState, scheduleGame.Status.StatusCode))
				return status
			}(),
			FetchedAt:     time.Now(),
			Period:        inning,
			PeriodType:    "INNING",
			TimeRemaining: countDisplay,
			Clock:         gameTimeDisplay,
			Venue:         venue,
			Details:       details,
			Statistics:    models.TeamStats{Hits: homeStats.Hits + awayStats.Hits, Errors: homeStats.Errors + awayStats.Errors},
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
	// Check for delays first, regardless of abstract game state
	if scheduleGame.Status.StatusCode == "IR" ||
		scheduleGame.Status.DetailedState == "Delayed" ||
		scheduleGame.Status.DetailedState == "Delayed: Rain" {
		logger.Info(fmt.Sprintf("Game %d mapped to StatusDelayed - StatusCode: %s, DetailedState: %s",
			scheduleGame.GamePk, scheduleGame.Status.StatusCode, scheduleGame.Status.DetailedState))
		return models.StatusDelayed
	}

	// Use StatusCode for more accurate status detection
	switch scheduleGame.Status.StatusCode {
	case "F": // Final
		logger.Info(fmt.Sprintf("Game %d mapped to StatusEnded - StatusCode: %s, AbstractGameState: %s",
			scheduleGame.GamePk, scheduleGame.Status.StatusCode, scheduleGame.Status.AbstractGameState))
		return models.StatusEnded
	case "S": // Scheduled/Preview
		logger.Info(fmt.Sprintf("Game %d mapped to StatusUpcoming - StatusCode: %s, AbstractGameState: %s",
			scheduleGame.GamePk, scheduleGame.Status.StatusCode, scheduleGame.Status.AbstractGameState))
		return models.StatusUpcoming
	case "L": // Live
		logger.Info(fmt.Sprintf("Game %d mapped to StatusActive - StatusCode: %s, AbstractGameState: %s",
			scheduleGame.GamePk, scheduleGame.Status.StatusCode, scheduleGame.Status.AbstractGameState))
		return models.StatusActive
	default:
		// Fallback to AbstractGameState if StatusCode is not recognized
		logger.Info(fmt.Sprintf("Game %d using fallback mapping - StatusCode: %s, AbstractGameState: %s",
			scheduleGame.GamePk, scheduleGame.Status.StatusCode, scheduleGame.Status.AbstractGameState))
		switch scheduleGame.Status.AbstractGameState {
		case STATUS_FINAL:
			logger.Info(fmt.Sprintf("Game %d fallback mapped to StatusEnded", scheduleGame.GamePk))
			return models.StatusEnded
		case STATUS_UPCOMING:
			logger.Info(fmt.Sprintf("Game %d fallback mapped to StatusUpcoming", scheduleGame.GamePk))
			return models.StatusUpcoming
		case STATUS_ACTIVE:
			logger.Info(fmt.Sprintf("Game %d fallback mapped to StatusActive", scheduleGame.GamePk))
			return models.StatusActive
		default:
			logger.Info(fmt.Sprintf("Game %d fallback mapped to StatusActive (default)", scheduleGame.GamePk))
			return models.StatusActive
		}
	}
}
func gameStatusFromStatusCode(statusCode string) models.GameStatus {
	switch statusCode {
	case "7":
		return models.StatusEnded
	case "IR": // In Progress - Rain Delay
		return models.StatusDelayed
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

// extractCurrentPitcher finds the current pitcher from the boxscore teams
func (s MLBService) extractCurrentPitcher(teams mlb.BoxscoreTeams) models.Player {
	// Check both home and away teams for current pitcher
	for _, team := range []mlb.TeamBoxscore{teams.Home, teams.Away} {
		for _, player := range team.Players {
			if player.GameStatus.IsCurrentPitcher {
				// Convert jersey number from string to int
				jerseyNumber := 0
				if player.JerseyNumber != "" {
					if num, err := strconv.Atoi(player.JerseyNumber); err == nil {
						jerseyNumber = num
					}
				}

				return models.Player{
					Id:       strconv.Itoa(player.Person.ID),
					Name:     player.Person.FullName,
					Number:   jerseyNumber,
					Position: player.Position.Name,
					Team: models.Team{
						ID:       team.Team.ID,
						TeamCode: "", // Will be filled by caller if needed
						TeamName: team.Team.Name,
						LeagueID: models.LeagueIdMLB,
					},
				}
			}
		}
	}
	return models.Player{} // Return empty player if none found
}

// extractCurrentBatter finds the current batter from the boxscore teams
func (s MLBService) extractCurrentBatter(teams mlb.BoxscoreTeams) models.Player {
	// Check both home and away teams for current batter
	for _, team := range []mlb.TeamBoxscore{teams.Home, teams.Away} {
		for _, player := range team.Players {
			if player.GameStatus.IsCurrentBatter {
				// Convert jersey number from string to int
				jerseyNumber := 0
				if player.JerseyNumber != "" {
					if num, err := strconv.Atoi(player.JerseyNumber); err == nil {
						jerseyNumber = num
					}
				}

				return models.Player{
					Id:       strconv.Itoa(player.Person.ID),
					Name:     player.Person.FullName,
					Number:   jerseyNumber,
					Position: player.Position.Name,
					Team: models.Team{
						ID:       team.Team.ID,
						TeamCode: "", // Will be filled by caller if needed
						TeamName: team.Team.Name,
						LeagueID: models.LeagueIdMLB,
					},
				}
			}
		}
	}
	return models.Player{} // Return empty player if none found
}
