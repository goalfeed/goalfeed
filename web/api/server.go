package webApi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	cflClients "goalfeed/clients/leagues/cfl"
	mlbClients "goalfeed/clients/leagues/mlb"
	nflClients "goalfeed/clients/leagues/nfl"
	nhlClients "goalfeed/clients/leagues/nhl"
	"goalfeed/config"
	"goalfeed/models"
	"goalfeed/services/leagues"
	cflServices "goalfeed/services/leagues/cfl"
	iihfServices "goalfeed/services/leagues/iihf"
	mlbServices "goalfeed/services/leagues/mlb"
	nflServices "goalfeed/services/leagues/nfl"
	nhlServices "goalfeed/services/leagues/nhl"
	"goalfeed/targets/memoryStore"
	"goalfeed/targets/notify"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for single-server mode
	},
}

type WebSocketHub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

var hub = &WebSocketHub{
	clients:    make(map[*websocket.Conn]bool),
	broadcast:  make(chan []byte),
	register:   make(chan *websocket.Conn),
	unregister: make(chan *websocket.Conn),
}

func (h *WebSocketHub) run() {
	for {
		select {
		case conn := <-h.register:
			h.clients[conn] = true
			// Send initial games list
			games := normalizeGamesData(memoryStore.GetAllGames())
			message := WebSocketMessage{
				Type: "games_list",
				Data: games,
			}
			if data, err := json.Marshal(message); err == nil {
				conn.WriteMessage(websocket.TextMessage, data)
			}
		case conn := <-h.unregister:
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				conn.Close()
			}
		case message := <-h.broadcast:
			for conn := range h.clients {
				err := conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					delete(h.clients, conn)
					conn.Close()
				}
			}
		}
	}
}

type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type ApiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

// normalizeGamesData ensures active games aren't marked as upcoming and persists changes
func normalizeGamesData(games []models.Game) []models.Game {
	for i := range games {
		if games[i].CurrentState.Status != models.StatusEnded {
			if games[i].CurrentState.Period > 0 || (games[i].CurrentState.Clock != "" && games[i].CurrentState.Clock != "TBD") {
				if games[i].CurrentState.Status != models.StatusActive {
					games[i].CurrentState.Status = models.StatusActive
					memoryStore.SetGame(games[i])
				}
			}
		}
	}
	return games
}

func BroadcastGameUpdate(game models.Game) {
	// Normalize single game before broadcasting
	if game.CurrentState.Status != models.StatusEnded {
		if game.CurrentState.Period > 0 || (game.CurrentState.Clock != "" && game.CurrentState.Clock != "TBD") {
			game.CurrentState.Status = models.StatusActive
			memoryStore.SetGame(game)
		}
	}
	message := WebSocketMessage{
		Type: "game_update",
		Data: game,
	}
	if data, err := json.Marshal(message); err == nil {
		hub.broadcast <- data
	}
}

func BroadcastEvent(event models.Event) {
	message := WebSocketMessage{
		Type: "event",
		Data: event,
	}
	if data, err := json.Marshal(message); err == nil {
		hub.broadcast <- data
	}
}

func BroadcastGamesList() {
	games := normalizeGamesData(memoryStore.GetAllGames())
	message := WebSocketMessage{
		Type: "games_list",
		Data: games,
	}
	if data, err := json.Marshal(message); err == nil {
		hub.broadcast <- data
	}
}

// Helper to check if a team is monitored given list (supports "*")
func isTeamMonitored(monitored []string, teamCode string) bool {
	for _, t := range monitored {
		if t == "*" || strings.EqualFold(t, teamCode) {
			return true
		}
	}
	return false
}

// Refresh active games based on current configuration
func refreshActiveGamesInternal() {
	type leagueCfg struct {
		id   models.League
		name string
	}
	leaguesToScan := []leagueCfg{
		{models.LeagueIdNHL, "nhl"},
		{models.LeagueIdMLB, "mlb"},
		{models.LeagueIdCFL, "cfl"},
		{models.LeagueIdIIHF, "iihf"},
		{models.LeagueIdNFL, "nfl"},
	}

	for _, lc := range leaguesToScan {
		monitored := config.GetStringSlice("watch." + lc.name)
		if len(monitored) == 0 {
			continue
		}

		var svc leagues.ILeagueService
		switch lc.id {
		case models.LeagueIdNHL:
			svc = nhlServices.NHLService{Client: nhlClients.NHLApiClient{}}
		case models.LeagueIdMLB:
			svc = mlbServices.MLBService{Client: mlbClients.MLBApiClient{}}
		case models.LeagueIdCFL:
			svc = cflServices.CFLService{Client: cflClients.CFLApiClient{}}
		case models.LeagueIdIIHF:
			svc = iihfServices.IIHFService{}
		case models.LeagueIdNFL:
			svc = nflServices.NFLService{Client: nflClients.NFLAPIClient{}}
		default:
			continue
		}

		ch := make(chan []models.Game)
		go svc.GetActiveGames(ch)
		active := <-ch

		// Track existing keys to avoid duplicates
		existing := make(map[string]bool)
		for _, k := range memoryStore.GetActiveGameKeys() {
			existing[k] = true
		}
		for _, g := range active {
			if isTeamMonitored(monitored, g.CurrentState.Home.Team.TeamCode) || isTeamMonitored(monitored, g.CurrentState.Away.Team.TeamCode) {
				key := g.GetGameKey()
				// Enforce active if period/clock indicate gameplay and not ended
				if g.CurrentState.Status != models.StatusEnded {
					if (g.CurrentState.Period > 0) || (g.CurrentState.Clock != "" && g.CurrentState.Clock != "TBD") {
						g.CurrentState.Status = models.StatusActive
					}
				}
				// Always set the latest game snapshot (updates status, clock, etc.)
				memoryStore.SetGame(g)
				if !existing[key] {
					memoryStore.AppendActiveGame(g)
					existing[key] = true
				}
			}
		}
	}

	// Broadcast updated list
	BroadcastGamesList()
}

func refreshActiveGames(c *gin.Context) {
	go refreshActiveGamesInternal()
	c.JSON(http.StatusOK, ApiResponse{Success: true, Message: "Refresh started"})
}

// DEBUG: Force-add an NFL game by ESPN event ID
func addNFLGame(c *gin.Context) {
	eventID := c.Query("event")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, ApiResponse{Success: false, Message: "event parameter is required"})
		return
	}
	svc := nflServices.NFLService{Client: nflClients.NFLAPIClient{}}
	game := svc.GameFromScoreboard(eventID)
	if game.GameCode == "" {
		c.JSON(http.StatusNotFound, ApiResponse{Success: false, Message: "Could not fetch game from summary"})
		return
	}
	memoryStore.SetGame(game)
	memoryStore.AppendActiveGame(game)
	BroadcastGameUpdate(game)
	BroadcastGamesList()
	c.JSON(http.StatusOK, ApiResponse{Success: true, Data: game, Message: "NFL game added"})
}

func handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	hub.register <- conn
	defer func() { hub.unregister <- conn }()

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func getGames(c *gin.Context) {
	games := normalizeGamesData(memoryStore.GetAllGames())
	// Enrich NFL games missing situation details
	for i := range games {
		g := &games[i]
		if g.LeagueId == models.LeagueIdNFL {
			// If any detail is missing, try to enrich from summary
			if g.CurrentState.Details.Down == 0 || g.CurrentState.Details.Distance == 0 || g.CurrentState.Details.Possession == "" || g.CurrentState.Details.YardLine == 0 {
				svc := nflServices.NFLService{Client: nflClients.NFLAPIClient{}}
				enriched := svc.GameFromScoreboard(g.GameCode)
				// Merge only missing fields
				if enriched.CurrentState.Details.Down > 0 && g.CurrentState.Details.Down == 0 {
					g.CurrentState.Details.Down = enriched.CurrentState.Details.Down
				}
				if enriched.CurrentState.Details.Distance > 0 && g.CurrentState.Details.Distance == 0 {
					g.CurrentState.Details.Distance = enriched.CurrentState.Details.Distance
				}
				if enriched.CurrentState.Details.Possession != "" && g.CurrentState.Details.Possession == "" {
					g.CurrentState.Details.Possession = enriched.CurrentState.Details.Possession
				}
				if enriched.CurrentState.Details.YardLine > 0 && g.CurrentState.Details.YardLine == 0 {
					g.CurrentState.Details.YardLine = enriched.CurrentState.Details.YardLine
				}
				// Apply halftime labeling if provided
				if enriched.CurrentState.PeriodType == "HALFTIME" {
					g.CurrentState.PeriodType = enriched.CurrentState.PeriodType
					g.CurrentState.Clock = enriched.CurrentState.Clock
				}
				memoryStore.SetGame(*g)
			}
		}
	}
	c.JSON(http.StatusOK, ApiResponse{
		Success: true,
		Data:    games,
	})
}

func getUpcomingGames(c *gin.Context) {
	// Get all league services
	var allUpcomingGames []models.Game

	// Define league configurations
	leagueConfigs := []struct {
		leagueId   models.League
		leagueName string
	}{
		{models.LeagueIdNHL, "nhl"},
		{models.LeagueIdMLB, "mlb"},
		{models.LeagueIdCFL, "cfl"},
		{models.LeagueIdIIHF, "iihf"},
		{models.LeagueIdNFL, "nfl"},
	}

	// Get timeframe filter (default to next 7 days)
	now := time.Now()
	oneWeekFromNow := now.AddDate(0, 0, 7)

	for _, leagueConfig := range leagueConfigs {
		// Get monitored teams for this league
		monitoredTeams := config.GetStringSlice("watch." + leagueConfig.leagueName)
		if len(monitoredTeams) == 0 {
			continue
		}

		// Get the league service for this league
		var leagueService leagues.ILeagueService
		switch leagueConfig.leagueId {
		case models.LeagueIdNHL:
			leagueService = nhlServices.NHLService{Client: nhlClients.NHLApiClient{}}
		case models.LeagueIdMLB:
			leagueService = mlbServices.MLBService{Client: mlbClients.MLBApiClient{}}
		case models.LeagueIdCFL:
			leagueService = cflServices.CFLService{Client: cflClients.CFLApiClient{}}
		case models.LeagueIdIIHF:
			leagueService = iihfServices.IIHFService{}
		case models.LeagueIdNFL:
			leagueService = nflServices.NFLService{Client: nflClients.NFLAPIClient{}}
		default:
			continue
		}

		// Get upcoming games for this league
		upcomingChan := make(chan []models.Game)
		go leagueService.GetUpcomingGames(upcomingChan)
		upcomingGames := <-upcomingChan

		// Filter to only include games with monitored teams and within timeframe
		for _, game := range upcomingGames {
			// Check if either team is being monitored
			isMonitored := false
			for _, teamCode := range monitoredTeams {
				// Handle wildcard "*" to include all teams
				if teamCode == "*" {
					isMonitored = true
					break
				}
				if game.CurrentState.Away.Team.TeamCode == teamCode || game.CurrentState.Home.Team.TeamCode == teamCode {
					isMonitored = true
					break
				}
			}

			// Check if game is within the next week
			isWithinTimeframe := true
			if !game.GameDetails.GameDate.IsZero() {
				isWithinTimeframe = game.GameDetails.GameDate.After(now) && game.GameDetails.GameDate.Before(oneWeekFromNow)
			}

			if isMonitored && isWithinTimeframe {
				allUpcomingGames = append(allUpcomingGames, game)
			}
		}
	}

	// Sort games by date
	sort.Slice(allUpcomingGames, func(i, j int) bool {
		if allUpcomingGames[i].GameDetails.GameDate.IsZero() || allUpcomingGames[j].GameDetails.GameDate.IsZero() {
			return false
		}
		return allUpcomingGames[i].GameDetails.GameDate.Before(allUpcomingGames[j].GameDetails.GameDate)
	})

	c.JSON(http.StatusOK, ApiResponse{
		Success: true,
		Data:    allUpcomingGames,
	})
}

func getLeagues(c *gin.Context) {
	leagues := []map[string]interface{}{
		{"leagueId": 1, "leagueName": "NHL", "teams": config.GetStringSlice("watch.nhl")},
		{"leagueId": 2, "leagueName": "MLB", "teams": config.GetStringSlice("watch.mlb")},
		{"leagueId": 5, "leagueName": "CFL", "teams": config.GetStringSlice("watch.cfl")},
		{"leagueId": 6, "leagueName": "NFL", "teams": config.GetStringSlice("watch.nfl")},
	}
	c.JSON(http.StatusOK, ApiResponse{
		Success: true,
		Data:    leagues,
	})
}

func updateLeagueConfig(c *gin.Context) {
	var config struct {
		LeagueId int      `json:"leagueId"`
		Teams    []string `json:"teams"`
	}
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Success: false,
			Message: "Invalid JSON",
		})
		return
	}

	// Update the configuration based on league ID
	var leagueKey string
	switch config.LeagueId {
	case 1:
		leagueKey = "watch.nhl"
	case 2:
		leagueKey = "watch.mlb"
	case 5:
		leagueKey = "watch.cfl"
	case 6:
		leagueKey = "watch.nfl"
	default:
		c.JSON(http.StatusBadRequest, ApiResponse{
			Success: false,
			Message: "Invalid league ID",
		})
		return
	}

	// Update the viper configuration
	viper.Set(leagueKey, config.Teams)

	// Write the configuration to file
	if err := viper.WriteConfig(); err != nil {
		log.Printf("Failed to write config: %v", err)
		c.JSON(http.StatusInternalServerError, ApiResponse{
			Success: false,
			Message: "Failed to save configuration",
		})
		return
	}

	// Trigger an immediate refresh in the background so active games are populated
	go refreshActiveGamesInternal()

	c.JSON(http.StatusOK, ApiResponse{
		Success: true,
		Message: "Configuration updated successfully",
	})
}

func getEvents(c *gin.Context) {
	// Return recent events
	events := []models.Event{} // TODO: Implement event storage
	c.JSON(http.StatusOK, ApiResponse{
		Success: true,
		Data:    events,
	})
}

func clearGames(c *gin.Context) {
	memoryStore.ClearAllGames()
	c.JSON(http.StatusOK, ApiResponse{
		Success: true,
		Message: "All games cleared from memory store",
	})
}

func getAllTeams(c *gin.Context) {
	leagueIdStr := c.Query("leagueId")
	if leagueIdStr == "" {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Success: false,
			Message: "leagueId parameter is required",
		})
		return
	}

	var leagueId int
	if _, err := fmt.Sscanf(leagueIdStr, "%d", &leagueId); err != nil {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Success: false,
			Message: "Invalid leagueId",
		})
		return
	}

	var teams []map[string]interface{}

	switch models.League(leagueId) {
	case models.LeagueIdNHL:
		// NHL teams with logos
		nhlTeams := []map[string]string{
			{"code": "ANA", "name": "Anaheim Ducks", "location": "Anaheim", "logo": "https://assets.nhle.com/logos/nhl/svg/ANA_light.svg"},
			{"code": "ARI", "name": "Arizona Coyotes", "location": "Arizona", "logo": "https://assets.nhle.com/logos/nhl/svg/ARI_light.svg"},
			{"code": "BOS", "name": "Boston Bruins", "location": "Boston", "logo": "https://assets.nhle.com/logos/nhl/svg/BOS_light.svg"},
			{"code": "BUF", "name": "Buffalo Sabres", "location": "Buffalo", "logo": "https://assets.nhle.com/logos/nhl/svg/BUF_light.svg"},
			{"code": "CGY", "name": "Calgary Flames", "location": "Calgary", "logo": "https://assets.nhle.com/logos/nhl/svg/CGY_light.svg"},
			{"code": "CAR", "name": "Carolina Hurricanes", "location": "Carolina", "logo": "https://assets.nhle.com/logos/nhl/svg/CAR_light.svg"},
			{"code": "CHI", "name": "Chicago Blackhawks", "location": "Chicago", "logo": "https://assets.nhle.com/logos/nhl/svg/CHI_light.svg"},
			{"code": "COL", "name": "Colorado Avalanche", "location": "Colorado", "logo": "https://assets.nhle.com/logos/nhl/svg/COL_light.svg"},
			{"code": "CBJ", "name": "Columbus Blue Jackets", "location": "Columbus", "logo": "https://assets.nhle.com/logos/nhl/svg/CBJ_light.svg"},
			{"code": "DAL", "name": "Dallas Stars", "location": "Dallas", "logo": "https://assets.nhle.com/logos/nhl/svg/DAL_light.svg"},
			{"code": "DET", "name": "Detroit Red Wings", "location": "Detroit", "logo": "https://assets.nhle.com/logos/nhl/svg/DET_light.svg"},
			{"code": "EDM", "name": "Edmonton Oilers", "location": "Edmonton", "logo": "https://assets.nhle.com/logos/nhl/svg/EDM_light.svg"},
			{"code": "FLA", "name": "Florida Panthers", "location": "Florida", "logo": "https://assets.nhle.com/logos/nhl/svg/FLA_light.svg"},
			{"code": "LAK", "name": "Los Angeles Kings", "location": "Los Angeles", "logo": "https://assets.nhle.com/logos/nhl/svg/LAK_light.svg"},
			{"code": "MIN", "name": "Minnesota Wild", "location": "Minnesota", "logo": "https://assets.nhle.com/logos/nhl/svg/MIN_light.svg"},
			{"code": "MTL", "name": "Montreal Canadiens", "location": "Montreal", "logo": "https://assets.nhle.com/logos/nhl/svg/MTL_light.svg"},
			{"code": "NSH", "name": "Nashville Predators", "location": "Nashville", "logo": "https://assets.nhle.com/logos/nhl/svg/NSH_light.svg"},
			{"code": "NJD", "name": "New Jersey Devils", "location": "New Jersey", "logo": "https://assets.nhle.com/logos/nhl/svg/NJD_light.svg"},
			{"code": "NYI", "name": "New York Islanders", "location": "New York", "logo": "https://assets.nhle.com/logos/nhl/svg/NYI_light.svg"},
			{"code": "NYR", "name": "New York Rangers", "location": "New York", "logo": "https://assets.nhle.com/logos/nhl/svg/NYR_light.svg"},
			{"code": "OTT", "name": "Ottawa Senators", "location": "Ottawa", "logo": "https://assets.nhle.com/logos/nhl/svg/OTT_light.svg"},
			{"code": "PHI", "name": "Philadelphia Flyers", "location": "Philadelphia", "logo": "https://assets.nhle.com/logos/nhl/svg/PHI_light.svg"},
			{"code": "PIT", "name": "Pittsburgh Penguins", "location": "Pittsburgh", "logo": "https://assets.nhle.com/logos/nhl/svg/PIT_light.svg"},
			{"code": "SJ", "name": "San Jose Sharks", "location": "San Jose", "logo": "https://assets.nhle.com/logos/nhl/svg/SJ_light.svg"},
			{"code": "SEA", "name": "Seattle Kraken", "location": "Seattle", "logo": "https://assets.nhle.com/logos/nhl/svg/SEA_light.svg"},
			{"code": "STL", "name": "St. Louis Blues", "location": "St. Louis", "logo": "https://assets.nhle.com/logos/nhl/svg/STL_light.svg"},
			{"code": "TB", "name": "Tampa Bay Lightning", "location": "Tampa Bay", "logo": "https://assets.nhle.com/logos/nhl/svg/TB_light.svg"},
			{"code": "TOR", "name": "Toronto Maple Leafs", "location": "Toronto", "logo": "https://assets.nhle.com/logos/nhl/svg/TOR_light.svg"},
			{"code": "VAN", "name": "Vancouver Canucks", "location": "Vancouver", "logo": "https://assets.nhle.com/logos/nhl/svg/VAN_light.svg"},
			{"code": "VGK", "name": "Vegas Golden Knights", "location": "Vegas", "logo": "https://assets.nhle.com/logos/nhl/svg/VGK_light.svg"},
			{"code": "WSH", "name": "Washington Capitals", "location": "Washington", "logo": "https://assets.nhle.com/logos/nhl/svg/WSH_light.svg"},
			{"code": "WPG", "name": "Winnipeg Jets", "location": "Winnipeg", "logo": "https://assets.nhle.com/logos/nhl/svg/WPG_light.svg"},
		}
		for _, team := range nhlTeams {
			teams = append(teams, map[string]interface{}{
				"code":     team["code"],
				"name":     team["name"],
				"location": team["location"],
				"logo":     team["logo"],
			})
		}
	case models.LeagueIdMLB:
		// MLB teams with logos
		mlbTeams := []map[string]string{
			{"code": "ARI", "name": "Arizona Diamondbacks", "location": "Arizona", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/ari.png"},
			{"code": "ATL", "name": "Atlanta Braves", "location": "Atlanta", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/atl.png"},
			{"code": "BAL", "name": "Baltimore Orioles", "location": "Baltimore", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/bal.png"},
			{"code": "BOS", "name": "Boston Red Sox", "location": "Boston", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/bos.png"},
			{"code": "CHC", "name": "Chicago Cubs", "location": "Chicago", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/chc.png"},
			{"code": "CWS", "name": "Chicago White Sox", "location": "Chicago", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/cws.png"},
			{"code": "CIN", "name": "Cincinnati Reds", "location": "Cincinnati", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/cin.png"},
			{"code": "CLE", "name": "Cleveland Guardians", "location": "Cleveland", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/cle.png"},
			{"code": "COL", "name": "Colorado Rockies", "location": "Colorado", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/col.png"},
			{"code": "DET", "name": "Detroit Tigers", "location": "Detroit", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/det.png"},
			{"code": "HOU", "name": "Houston Astros", "location": "Houston", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/hou.png"},
			{"code": "KC", "name": "Kansas City Royals", "location": "Kansas City", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/kc.png"},
			{"code": "LAA", "name": "Los Angeles Angels", "location": "Los Angeles", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/laa.png"},
			{"code": "LAD", "name": "Los Angeles Dodgers", "location": "Los Angeles", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/lad.png"},
			{"code": "MIA", "name": "Miami Marlins", "location": "Miami", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/mia.png"},
			{"code": "MIL", "name": "Milwaukee Brewers", "location": "Milwaukee", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/mil.png"},
			{"code": "MIN", "name": "Minnesota Twins", "location": "Minnesota", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/min.png"},
			{"code": "NYM", "name": "New York Mets", "location": "New York", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/nym.png"},
			{"code": "NYY", "name": "New York Yankees", "location": "New York", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/nyy.png"},
			{"code": "OAK", "name": "Oakland Athletics", "location": "Oakland", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/oak.png"},
			{"code": "PHI", "name": "Philadelphia Phillies", "location": "Philadelphia", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/phi.png"},
			{"code": "PIT", "name": "Pittsburgh Pirates", "location": "Pittsburgh", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/pit.png"},
			{"code": "SD", "name": "San Diego Padres", "location": "San Diego", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/sd.png"},
			{"code": "SF", "name": "San Francisco Giants", "location": "San Francisco", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/sf.png"},
			{"code": "SEA", "name": "Seattle Mariners", "location": "Seattle", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/sea.png"},
			{"code": "STL", "name": "St. Louis Cardinals", "location": "St. Louis", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/stl.png"},
			{"code": "TB", "name": "Tampa Bay Rays", "location": "Tampa Bay", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/tb.png"},
			{"code": "TEX", "name": "Texas Rangers", "location": "Texas", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/tex.png"},
			{"code": "TOR", "name": "Toronto Blue Jays", "location": "Toronto", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/tor.png"},
			{"code": "WSH", "name": "Washington Nationals", "location": "Washington", "logo": "https://a.espncdn.com/i/teamlogos/mlb/500/wsh.png"},
		}
		for _, team := range mlbTeams {
			teams = append(teams, map[string]interface{}{
				"code":     team["code"],
				"name":     team["name"],
				"location": team["location"],
				"logo":     team["logo"],
			})
		}
	case models.LeagueIdNFL:
		// NFL teams with logos
		nflTeams := []map[string]string{
			{"code": "ARI", "name": "Arizona Cardinals", "location": "Arizona", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/ari.png"},
			{"code": "ATL", "name": "Atlanta Falcons", "location": "Atlanta", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/atl.png"},
			{"code": "BAL", "name": "Baltimore Ravens", "location": "Baltimore", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/bal.png"},
			{"code": "BUF", "name": "Buffalo Bills", "location": "Buffalo", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/buf.png"},
			{"code": "CAR", "name": "Carolina Panthers", "location": "Carolina", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/car.png"},
			{"code": "CHI", "name": "Chicago Bears", "location": "Chicago", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/chi.png"},
			{"code": "CIN", "name": "Cincinnati Bengals", "location": "Cincinnati", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/cin.png"},
			{"code": "CLE", "name": "Cleveland Browns", "location": "Cleveland", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/cle.png"},
			{"code": "DAL", "name": "Dallas Cowboys", "location": "Dallas", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/dal.png"},
			{"code": "DEN", "name": "Denver Broncos", "location": "Denver", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/den.png"},
			{"code": "DET", "name": "Detroit Lions", "location": "Detroit", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/det.png"},
			{"code": "GB", "name": "Green Bay Packers", "location": "Green Bay", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/gb.png"},
			{"code": "HOU", "name": "Houston Texans", "location": "Houston", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/hou.png"},
			{"code": "IND", "name": "Indianapolis Colts", "location": "Indianapolis", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/ind.png"},
			{"code": "JAX", "name": "Jacksonville Jaguars", "location": "Jacksonville", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/jax.png"},
			{"code": "KC", "name": "Kansas City Chiefs", "location": "Kansas City", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/kc.png"},
			{"code": "LV", "name": "Las Vegas Raiders", "location": "Las Vegas", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/lv.png"},
			{"code": "LAC", "name": "Los Angeles Chargers", "location": "Los Angeles", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/lac.png"},
			{"code": "LAR", "name": "Los Angeles Rams", "location": "Los Angeles", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/lar.png"},
			{"code": "MIA", "name": "Miami Dolphins", "location": "Miami", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/mia.png"},
			{"code": "MIN", "name": "Minnesota Vikings", "location": "Minnesota", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/min.png"},
			{"code": "NE", "name": "New England Patriots", "location": "New England", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/ne.png"},
			{"code": "NO", "name": "New Orleans Saints", "location": "New Orleans", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/no.png"},
			{"code": "NYG", "name": "New York Giants", "location": "New York", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/nyg.png"},
			{"code": "NYJ", "name": "New York Jets", "location": "New York", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/nyj.png"},
			{"code": "PHI", "name": "Philadelphia Eagles", "location": "Philadelphia", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/phi.png"},
			{"code": "PIT", "name": "Pittsburgh Steelers", "location": "Pittsburgh", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/pit.png"},
			{"code": "SF", "name": "San Francisco 49ers", "location": "San Francisco", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/sf.png"},
			{"code": "SEA", "name": "Seattle Seahawks", "location": "Seattle", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/sea.png"},
			{"code": "TB", "name": "Tampa Bay Buccaneers", "location": "Tampa Bay", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/tb.png"},
			{"code": "TEN", "name": "Tennessee Titans", "location": "Tennessee", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/ten.png"},
			{"code": "WSH", "name": "Washington Commanders", "location": "Washington", "logo": "https://a.espncdn.com/i/teamlogos/nfl/500/wsh.png"},
		}
		for _, team := range nflTeams {
			teams = append(teams, map[string]interface{}{
				"code":     team["code"],
				"name":     team["name"],
				"location": team["location"],
				"logo":     team["logo"],
			})
		}
	case models.LeagueIdCFL:
		// CFL teams - logos will be handled by frontend fallback
		cflTeams := []map[string]string{
			{"code": "BC", "name": "BC Lions", "location": "Vancouver"},
			{"code": "CGY", "name": "Calgary Stampeders", "location": "Calgary"},
			{"code": "EDM", "name": "Edmonton Elks", "location": "Edmonton"},
			{"code": "HAM", "name": "Hamilton Tiger-Cats", "location": "Hamilton"},
			{"code": "MTL", "name": "Montreal Alouettes", "location": "Montreal"},
			{"code": "OTT", "name": "Ottawa Redblacks", "location": "Ottawa"},
			{"code": "SSK", "name": "Saskatchewan Roughriders", "location": "Saskatchewan"},
			{"code": "TOR", "name": "Toronto Argonauts", "location": "Toronto"},
			{"code": "WPG", "name": "Winnipeg Blue Bombers", "location": "Winnipeg"},
		}

		// Get monitored teams for CFL
		monitoredTeams := config.GetStringSlice("watch.cfl")

		for _, team := range cflTeams {
			// Check if this team is being monitored
			isMonitored := false
			for _, monitoredTeam := range monitoredTeams {
				if monitoredTeam == "*" || strings.EqualFold(monitoredTeam, team["code"]) {
					isMonitored = true
					break
				}
			}

			// Only add teams that are being monitored
			if isMonitored {
				teams = append(teams, map[string]interface{}{
					"code":     team["code"],
					"name":     team["name"],
					"location": team["location"],
					"logo":     "", // Empty logo - frontend will show team code as fallback
				})
			}
		}
	}

	c.JSON(http.StatusOK, ApiResponse{
		Success: true,
		Data:    teams,
	})
}

func buildFrontend() error {
	frontendDir := "./web/frontend"

	// Check if package.json exists
	if _, err := os.Stat(filepath.Join(frontendDir, "package.json")); os.IsNotExist(err) {
		return fmt.Errorf("package.json not found in %s", frontendDir)
	}

	// Check if npm is available
	if _, err := exec.LookPath("npm"); err != nil {
		return fmt.Errorf("npm not found. Please install Node.js and npm")
	}

	// Check if build directory already exists
	buildDir := filepath.Join(frontendDir, "build")
	if _, err := os.Stat(buildDir); os.IsNotExist(err) {
		// Run npm install
		installCmd := exec.Command("npm", "install")
		installCmd.Dir = frontendDir
		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("npm install failed: %v", err)
		}

		// Run npm run build
		buildCmd := exec.Command("npm", "run", "build")
		buildCmd.Dir = frontendDir
		if err := buildCmd.Run(); err != nil {
			return fmt.Errorf("npm run build failed: %v", err)
		}
	}

	return nil
}

func StartWebServer(port string) {
	// Start WebSocket hub
	go hub.run()

	// Expose broadcast functions for other packages to avoid import cycles
	notify.BroadcastGame = BroadcastGameUpdate
	notify.BroadcastGamesList = BroadcastGamesList

	// Try to build frontend
	if err := buildFrontend(); err != nil {
		log.Printf("Frontend build failed: %v", err)
		log.Println("Serving API-only mode. Install Node.js and npm to enable the web interface.")
	}

	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
	}))

	// API routes
	api := r.Group("/api")
	{
		api.GET("/games", getGames)
		api.GET("/upcoming", getUpcomingGames)
		api.GET("/leagues", getLeagues)
		api.POST("/leagues", updateLeagueConfig)
		api.POST("/refresh", refreshActiveGames)
		api.POST("/debug/nfl/add", addNFLGame)
		api.GET("/events", getEvents)
		api.GET("/teams", getAllTeams)
		api.POST("/clear", clearGames)
	}

	// WebSocket endpoint
	r.GET("/ws", handleWebSocket)

	// Serve static files
	frontendDir := "./web/frontend/build"
	if _, err := os.Stat(frontendDir); err == nil {
		r.Static("/static", filepath.Join(frontendDir, "static"))
		r.StaticFile("/", filepath.Join(frontendDir, "index.html"))
		r.NoRoute(func(c *gin.Context) {
			c.File(filepath.Join(frontendDir, "index.html"))
		})
	} else {
		// Fallback to basic HTML or API-only response
		r.GET("/", func(c *gin.Context) {
			fallbackPath := filepath.Join("./web/frontend", "public", "fallback.html")
			if _, err := os.Stat(fallbackPath); err == nil {
				c.File(fallbackPath)
			} else {
				c.JSON(http.StatusOK, gin.H{
					"api":     "/api",
					"message": "Goalfeed API Server",
					"status":  "running",
					"ws":      "/ws",
					"note":    "Frontend not available. Install Node.js and npm to enable the web interface.",
				})
			}
		})
	}

	log.Printf("Starting web server on port %s", port)
	r.Run(":" + port)
}
