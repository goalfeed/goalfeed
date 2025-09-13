package nfl

import (
	"goalfeed/clients/leagues/nfl"
	"goalfeed/models"
	"goalfeed/utils"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	STATUS_UPCOMING = "pre"
	STATUS_ACTIVE   = "in"
	STATUS_FINAL    = "post"
)

type NFLService struct {
	Client nfl.INFLAPIClient
}

var logger = utils.GetLogger()

func (s NFLService) GetLeagueName() string {
	return "NFL"
}

func (s NFLService) getSchedule() nfl.NFLScheduleResponse {
	return s.Client.GetNFLSchedule()
}

func (s NFLService) GetActiveGames(ret chan []models.Game) {
	schedule := s.getSchedule()
	var activeGames []models.Game

	for _, event := range schedule.Events {
		if gameStatusFromEvent(event) == models.StatusActive {
			// Use schedule data to ensure game appears immediately
			g := s.gameFromEvent(event)
			// Hydrate initial score/clock/state from scoreboard so we don't start at 0-0
			if event.ID != "" {
				fresh := s.GameFromScoreboard(event.ID)
				if fresh.GameCode != "" {
					// Merge critical runtime fields
					g.CurrentState.Home.Score = fresh.CurrentState.Home.Score
					g.CurrentState.Away.Score = fresh.CurrentState.Away.Score
					if fresh.CurrentState.Clock != "" {
						g.CurrentState.Clock = fresh.CurrentState.Clock
					}
					if fresh.CurrentState.Period > 0 {
						g.CurrentState.Period = fresh.CurrentState.Period
					}
					// Preserve Active if either source says Active
					if fresh.CurrentState.Status == models.StatusActive || g.CurrentState.Status == models.StatusActive {
						g.CurrentState.Status = models.StatusActive
					} else {
						g.CurrentState.Status = fresh.CurrentState.Status
					}
					// If halftime signaled in fresh, carry over labels
					if fresh.CurrentState.PeriodType == "HALFTIME" {
						g.CurrentState.PeriodType = fresh.CurrentState.PeriodType
						g.CurrentState.Clock = fresh.CurrentState.Clock
					}
					// Backfill team identifiers if any are missing from schedule
					if g.CurrentState.Home.Team.TeamCode == "" {
						g.CurrentState.Home.Team = fresh.CurrentState.Home.Team
					}
					if g.CurrentState.Away.Team.TeamCode == "" {
						g.CurrentState.Away.Team = fresh.CurrentState.Away.Team
					}
				}
			}
			activeGames = append(activeGames, g)
		}
	}
	ret <- activeGames
}

func (s NFLService) GetUpcomingGames(ret chan []models.Game) {
	schedule := s.getSchedule()
	var upcomingGames []models.Game

	for _, event := range schedule.Events {
		if gameStatusFromEvent(event) == models.StatusUpcoming {
			upcomingGames = append(upcomingGames, s.gameFromEvent(event))
		}
	}
	ret <- upcomingGames
}

func (s NFLService) GetGameUpdate(game models.Game, ret chan models.GameUpdate) {
	s.getGameUpdateFromScoreboard(game, ret)
}

func (s NFLService) getGameUpdateFromScoreboard(game models.Game, ret chan models.GameUpdate) {
	scoreboard := s.Client.GetNFLScoreBoard(game.GameCode)

	// Extract game info from scoreboard
	var newState models.GameState = game.CurrentState
	if len(scoreboard.Events) > 0 {
		event := scoreboard.Events[0]
		if len(event.Competitions) > 0 {
			competition := event.Competitions[0]
			if len(competition.Competitors) >= 2 {
				var awayTeam, homeTeam nfl.NFLScoreboardCompetitor

				// Find home and away teams by checking HomeAway field
				for _, competitor := range competition.Competitors {
					if competitor.HomeAway == "home" {
						homeTeam = competitor
					} else if competitor.HomeAway == "away" {
						awayTeam = competitor
					}
				}

				// Fallback: if API returned empty team data, reuse existing game state team info
				if homeTeam.Team.Abbreviation == "" && game.CurrentState.Home.Team.TeamCode != "" {
					homeTeam.Team.DisplayName = game.CurrentState.Home.Team.TeamName
					homeTeam.Team.Abbreviation = game.CurrentState.Home.Team.TeamCode
					homeTeam.Team.ID = game.CurrentState.Home.Team.ExtID
					homeTeam.Team.Logo = game.CurrentState.Home.Team.LogoURL
				}
				if awayTeam.Team.Abbreviation == "" && game.CurrentState.Away.Team.TeamCode != "" {
					awayTeam.Team.DisplayName = game.CurrentState.Away.Team.TeamName
					awayTeam.Team.Abbreviation = game.CurrentState.Away.Team.TeamCode
					awayTeam.Team.ID = game.CurrentState.Away.Team.ExtID
					awayTeam.Team.Logo = game.CurrentState.Away.Team.LogoURL
				}

				awayScore, _ := strconv.Atoi(awayTeam.Score)
				homeScore, _ := strconv.Atoi(homeTeam.Score)

				// Derive status directly from summary signals
				var derivedStatus models.GameStatus = models.StatusUpcoming
				if event.Status.Type.Completed {
					derivedStatus = models.StatusEnded
				} else if event.Status.DisplayClock != "" || event.Status.Period > 0 {
					derivedStatus = models.StatusActive
				}

				newState = models.GameState{
					Home: models.TeamState{
						Team:  s.teamFromCompetitor(homeTeam),
						Score: homeScore,
					},
					Away: models.TeamState{
						Team:  s.teamFromCompetitor(awayTeam),
						Score: awayScore,
					},
					Status:     derivedStatus,
					Period:     event.Status.Period,
					PeriodType: "QUARTER",
					Clock:      event.Status.DisplayClock,
					Venue: models.Venue{
						Name: competition.Venue.FullName,
					},
					Details: models.EventDetails{
						Possession: func() string {
							pt := scoreboard.Drives.Current.Start.PossessionText
							if pt == "" {
								// Fallback to current drive team abbreviation
								abbr := scoreboard.Drives.Current.Team.Abbreviation
								return strings.ToUpper(strings.TrimSpace(abbr))
							}
							parts := strings.Fields(pt)
							if len(parts) > 0 {
								return parts[0]
							}
							return ""
						}(),
						YardLine: scoreboard.Drives.Current.Start.YardLine,
						Down:     scoreboard.Drives.Current.Start.Down,
						Distance: scoreboard.Drives.Current.Start.Distance,
					},
				}

				// Label halftime when appropriate
				if strings.Contains(strings.ToLower(event.Status.Type.ShortDetail), "halftime") ||
					(newState.Period == 2 && newState.Clock == "0:00") {
					newState.PeriodType = "HALFTIME"
					newState.Clock = "HALFTIME"
				}

				// If situation is missing, attempt to parse from ShortDetail (e.g., "1st & 10 at CLE 25")
				if newState.Details.Down == 0 || newState.Details.Distance == 0 || newState.Details.Possession == "" || newState.Details.YardLine == 0 {
					d, dist, poss, yl := parseSituationShortDetail(event.Status.Type.ShortDetail)
					if newState.Details.Down == 0 && d > 0 {
						newState.Details.Down = d
					}
					if newState.Details.Distance == 0 && dist > 0 {
						newState.Details.Distance = dist
					}
					if newState.Details.Possession == "" && poss != "" {
						newState.Details.Possession = poss
					}
					if newState.Details.YardLine == 0 && yl > 0 {
						newState.Details.YardLine = yl
					}
				}
			}
		}
	}

	ret <- models.GameUpdate{
		OldState: game.CurrentState,
		NewState: newState,
	}
}

// parseSituationShortDetail parses strings like "1st & 10 at CLE 25" into components
func parseSituationShortDetail(s string) (down int, distance int, possession string, yardLine int) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, 0, "", 0
	}
	// Regex: <1st|2nd|3rd|4th> & <distance> at <TEAM> <yard>
	re := regexp.MustCompile(`^(?i)(1st|2nd|3rd|4th)\s*&\s*(\d+)(?:\s+at\s+([A-Z]{2,4})\s+(\d+))?`)
	m := re.FindStringSubmatch(s)
	if len(m) == 5 {
		// Map down text to number
		var d int
		switch strings.ToLower(m[1]) {
		case "1st":
			d = 1
		case "2nd":
			d = 2
		case "3rd":
			d = 3
		case "4th":
			d = 4
		}
		dist, _ := strconv.Atoi(m[2])
		team := strings.ToUpper(m[3])
		yl, _ := strconv.Atoi(m[4])
		return d, dist, team, yl
	}
	return 0, 0, "", 0
}

func (s NFLService) teamFromCompetitor(competitor nfl.NFLScoreboardCompetitor) models.Team {
	return models.Team{
		TeamName: competitor.Team.DisplayName,
		TeamCode: competitor.Team.Abbreviation,
		ExtID:    competitor.Team.ID,
		LeagueID: models.LeagueIdNFL,
		LogoURL:  competitor.Team.Logo,
	}
}

func (s NFLService) gameFromEvent(event nfl.NFLScheduleEvent) models.Game {
	var homeTeam, awayTeam nfl.NFLCompetitor
	var venue models.Venue
	var gameDate time.Time

	if len(event.Competitions) > 0 {
		competition := event.Competitions[0]
		venue = models.Venue{
			Name:  competition.Venue.FullName,
			City:  competition.Venue.Address.City,
			State: competition.Venue.Address.State,
		}

		if len(competition.Competitors) >= 2 {
			// Find home and away teams by checking HomeAway field
			for _, competitor := range competition.Competitors {
				if competitor.HomeAway == "home" {
					homeTeam = competitor
				} else if competitor.HomeAway == "away" {
					awayTeam = competitor
				}
			}
		}
	}

	// If team data is missing from schedule, hydrate from scoreboard summary
	if (homeTeam.Team.Abbreviation == "" || awayTeam.Team.Abbreviation == "") && event.ID != "" {
		summary := s.Client.GetNFLScoreBoard(event.ID)
		if len(summary.Events) > 0 && len(summary.Events[0].Competitions) > 0 {
			sc := summary.Events[0].Competitions[0]
			var sHome nfl.NFLScoreboardCompetitor
			var sAway nfl.NFLScoreboardCompetitor
			for _, comp := range sc.Competitors {
				if comp.HomeAway == "home" {
					sHome = comp
				} else if comp.HomeAway == "away" {
					sAway = comp
				}
			}
			// Copy team identifiers if missing
			if homeTeam.Team.Abbreviation == "" {
				homeTeam.Team.DisplayName = sHome.Team.DisplayName
				homeTeam.Team.Abbreviation = sHome.Team.Abbreviation
				homeTeam.Team.ID = sHome.Team.ID
				homeTeam.Team.Logo = sHome.Team.Logo
			}
			if awayTeam.Team.Abbreviation == "" {
				awayTeam.Team.DisplayName = sAway.Team.DisplayName
				awayTeam.Team.Abbreviation = sAway.Team.Abbreviation
				awayTeam.Team.ID = sAway.Team.ID
				awayTeam.Team.Logo = sAway.Team.Logo
			}
		}
	}

	// Parse the game date with multiple layouts
	if event.Date != "" {
		layouts := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05-07:00",
			"2006-01-02T15:04Z",
			"2006-01-02T15:04-07:00",
		}
		for _, layout := range layouts {
			if parsedDate, err := time.Parse(layout, event.Date); err == nil {
				gameDate = parsedDate
				break
			}
		}
	}

	// Fallback to competition date if event.Date is empty or failed parsing
	if gameDate.IsZero() && len(event.Competitions) > 0 {
		competition := event.Competitions[0]
		if competition.Date != "" {
			layouts := []string{
				time.RFC3339,
				time.RFC3339Nano,
				"2006-01-02T15:04:05Z",
				"2006-01-02T15:04:05-07:00",
				"2006-01-02T15:04Z",
				"2006-01-02T15:04-07:00",
			}
			for _, layout := range layouts {
				if parsedDate, err := time.Parse(layout, competition.Date); err == nil {
					gameDate = parsedDate
					break
				}
			}
		}
	}

	// Final fallback: fetch scoreboard summary to get reliable date
	if gameDate.IsZero() && event.ID != "" {
		summary := s.Client.GetNFLScoreBoard(event.ID)
		if len(summary.Events) > 0 {
			se := summary.Events[0]
			// Try event date first
			layouts := []string{
				time.RFC3339,
				time.RFC3339Nano,
				"2006-01-02T15:04:05Z",
				"2006-01-02T15:04:05-07:00",
				"2006-01-02T15:04Z",
				"2006-01-02T15:04-07:00",
			}
			if se.Date != "" {
				for _, layout := range layouts {
					if parsedDate, err := time.Parse(layout, se.Date); err == nil {
						gameDate = parsedDate
						break
					}
				}
			}
			// Then try competition date from summary
			if gameDate.IsZero() && len(se.Competitions) > 0 {
				cd := se.Competitions[0].Date
				if cd != "" {
					for _, layout := range layouts {
						if parsedDate, err := time.Parse(layout, cd); err == nil {
							gameDate = parsedDate
							break
						}
					}
				}
			}
		}
	}

	awayScore, _ := strconv.Atoi(awayTeam.Score)
	homeScore, _ := strconv.Atoi(homeTeam.Score)

	// Format game time for upcoming games
	var gameTimeDisplay string
	if event.Status.Type.State == STATUS_UPCOMING {
		if !gameDate.IsZero() {
			gameTimeDisplay = gameDate.Format("Mon 3:04 PM")
		} else {
			gameTimeDisplay = "TBD"
		}
	} else {
		gameTimeDisplay = event.Status.DisplayClock
	}

	// Precompute display time for details
	gameTimeStr := "TBD"
	if !gameDate.IsZero() {
		gameTimeStr = gameDate.Format("3:04 PM")
	}

	g := models.Game{
		CurrentState: models.GameState{
			ExtTimestamp: event.Date,
			Home: models.TeamState{
				Team:  s.teamFromScheduleCompetitor(homeTeam),
				Score: homeScore,
			},
			Away: models.TeamState{
				Team:  s.teamFromScheduleCompetitor(awayTeam),
				Score: awayScore,
			},
			Status: func() models.GameStatus {
				st := gameStatusFromEvent(event)
				// If schedule shows clock or period, force Active
				if st == models.StatusUpcoming && (event.Status.DisplayClock != "" || event.Status.Period > 0) {
					return models.StatusActive
				}
				return st
			}(),
			FetchedAt:  time.Now(),
			Period:     event.Status.Period,
			PeriodType: "QUARTER",
			Clock:      gameTimeDisplay,
			Venue:      venue,
		},
		GameCode:     event.ID,
		ExtTimestamp: event.Date,
		LeagueId:     models.LeagueIdNFL,
		GameDetails: models.GameDetails{
			GameId:     event.ID,
			Season:     strconv.Itoa(event.Season.Year),
			SeasonType: "REGULAR",
			Week:       event.Week.Number,
			GameDate:   gameDate,
			GameTime:   gameTimeStr,
			Timezone:   "UTC",
		},
	}
	if strings.Contains(strings.ToLower(event.Status.Type.ShortDetail), "halftime") || (g.CurrentState.Period == 2 && (event.Status.DisplayClock == "0:00" || g.CurrentState.Clock == "0:00")) {
		g.CurrentState.PeriodType = "HALFTIME"
		g.CurrentState.Clock = "HALFTIME"
	}
	return g
}

// GameFromScoreboard builds a Game using the summary endpoint (eventID)
func (s NFLService) GameFromScoreboard(eventID string) models.Game {
	sb := s.Client.GetNFLScoreBoard(eventID)
	if len(sb.Events) == 0 || len(sb.Events[0].Competitions) == 0 {
		// Fallback: minimal game using IDs only
		return models.Game{
			GameCode: eventID,
			LeagueId: models.LeagueIdNFL,
		}
	}
	ev := sb.Events[0]
	comp := ev.Competitions[0]
	var awayC, homeC nfl.NFLScoreboardCompetitor
	for _, c := range comp.Competitors {
		if c.HomeAway == "home" {
			homeC = c
		} else if c.HomeAway == "away" {
			awayC = c
		}
	}
	awayScore, _ := strconv.Atoi(awayC.Score)
	homeScore, _ := strconv.Atoi(homeC.Score)

	// Derive status
	var st models.GameStatus = models.StatusUpcoming
	if ev.Status.Type.Completed {
		st = models.StatusEnded
	} else if ev.Status.DisplayClock != "" || ev.Status.Period > 0 {
		st = models.StatusActive
	}

	// Period/clock
	period := ev.Status.Period
	clock := ev.Status.DisplayClock

	// Situation
	details := models.EventDetails{
		Possession: func() string {
			pt := sb.Drives.Current.Start.PossessionText
			if pt == "" {
				abbr := sb.Drives.Current.Team.Abbreviation
				return strings.ToUpper(strings.TrimSpace(abbr))
			}
			parts := strings.Fields(pt)
			if len(parts) > 0 {
				return parts[0]
			}
			return ""
		}(),
		YardLine: sb.Drives.Current.Start.YardLine,
		Down:     sb.Drives.Current.Start.Down,
		Distance: sb.Drives.Current.Start.Distance,
	}
	if details.Down == 0 && details.Distance == 0 && details.Possession == "" && details.YardLine == 0 {
		d, dist, poss, yl := parseSituationShortDetail(ev.Status.Type.ShortDetail)
		details.Down = d
		details.Distance = dist
		details.Possession = poss
		details.YardLine = yl
		// If still missing down/distance, try compact shortDownDistanceText (e.g., "2nd & 8")
		if (details.Down == 0 || details.Distance == 0) && sb.Drives.Current.Start.ShortDownDistanceText != "" {
			d2, dist2, _, _ := parseSituationShortDetail(sb.Drives.Current.Start.ShortDownDistanceText)
			if details.Down == 0 && d2 > 0 {
				details.Down = d2
			}
			if details.Distance == 0 && dist2 > 0 {
				details.Distance = dist2
			}
		}
	}

	g := models.Game{
		CurrentState: models.GameState{
			Home:       models.TeamState{Team: s.teamFromCompetitor(homeC), Score: homeScore},
			Away:       models.TeamState{Team: s.teamFromCompetitor(awayC), Score: awayScore},
			Status:     st,
			FetchedAt:  time.Now(),
			Period:     period,
			PeriodType: "QUARTER",
			Clock:      clock,
			Venue:      models.Venue{Name: comp.Venue.FullName},
			Details:    details,
		},
		GameCode: eventID,
		LeagueId: models.LeagueIdNFL,
		GameDetails: models.GameDetails{
			GameId:     eventID,
			Season:     strconv.Itoa(ev.Season.Year),
			SeasonType: "REGULAR",
			Week:       ev.Week.Number,
			GameDate:   time.Now(),
			GameTime:   "",
			Timezone:   "UTC",
		},
	}
	if strings.Contains(strings.ToLower(ev.Status.Type.ShortDetail), "halftime") || (g.CurrentState.Period == 2 && g.CurrentState.Clock == "0:00") {
		g.CurrentState.PeriodType = "HALFTIME"
		g.CurrentState.Clock = "HALFTIME"
	}
	return g
}

func (s NFLService) teamFromScheduleCompetitor(competitor nfl.NFLCompetitor) models.Team {
	return models.Team{
		TeamName: competitor.Team.DisplayName,
		TeamCode: competitor.Team.Abbreviation,
		ExtID:    competitor.Team.ID,
		LeagueID: models.LeagueIdNFL,
		LogoURL:  competitor.Team.Logo,
	}
}

func gameStatusFromEvent(event nfl.NFLScheduleEvent) models.GameStatus {
	return gameStatusFromEventStatus(event.Status)
}

func gameStatusFromEventStatus(status struct {
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
}) models.GameStatus {
	switch status.Type.State {
	case STATUS_FINAL:
		return models.StatusEnded
	case STATUS_UPCOMING:
		return models.StatusUpcoming
	case STATUS_ACTIVE:
		return models.StatusActive
	default:
		return models.StatusUpcoming
	}
}

func (s NFLService) GetEvents(update models.GameUpdate, ret chan []models.Event) {
	events := append(
		s.getTouchdownEvents(update.OldState.Home, update.NewState.Home, update.OldState.Away.Team),
		s.getTouchdownEvents(update.OldState.Away, update.NewState.Away, update.OldState.Home.Team)...,
	)
	ret <- events
}

func (s NFLService) getTouchdownEvents(oldState models.TeamState, newState models.TeamState, opponent models.Team) []models.Event {
	events := []models.Event{}
	diff := newState.Score - oldState.Score
	team := newState.Team

	// In NFL, we track touchdowns (7 points) and field goals (3 points)
	// For simplicity, we'll create events for each point scored
	for i := 0; i < diff; i++ {
		events = append(events, models.Event{
			TeamCode:     team.TeamCode,
			TeamName:     team.TeamName,
			TeamHash:     team.GetTeamHash(),
			LeagueId:     models.LeagueIdNFL,
			LeagueName:   s.GetLeagueName(),
			OpponentCode: opponent.TeamCode,
			OpponentName: opponent.TeamName,
			OpponentHash: opponent.GetTeamHash(),
		})
	}
	return events
}
