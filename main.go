package main

import (
	"fmt"
	cflClients "goalfeed/clients/leagues/cfl"
	mlbClients "goalfeed/clients/leagues/mlb"
	nflClients "goalfeed/clients/leagues/nfl"
	nhlClients "goalfeed/clients/leagues/nhl"
	"goalfeed/config"
	"goalfeed/models"
	"goalfeed/services/leagues"
	"goalfeed/services/leagues/cfl"
	"goalfeed/services/leagues/mlb"
	"goalfeed/services/leagues/nfl"
	"goalfeed/services/leagues/nhl"
	"goalfeed/targets/homeassistant"
	"goalfeed/targets/memoryStore"
	"goalfeed/utils"
	webApi "goalfeed/web/api"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/joho/godotenv"
)

var version string // This will be populated by the build process

var rootCmd = &cobra.Command{
	Use:   "goalfeed",
	Short: "Goalfeed main application",
	Long:  `Starts the Goalfeed application.`,
	Run: func(cmd *cobra.Command, args []string) {
		initialize()
		if viper.GetBool("web") {
			runWebMode()
		} else {
			runTickers()
		}
	},
}
var (
	leagueServices                    = map[int]leagues.ILeagueService{}
	needRefresh                       = false
	logger                            = utils.GetLogger()
	eventSender    func(models.Event) = homeassistant.SendEvent // Allow this to be replaced in tests
)

func init() {
	_ = godotenv.Load()
	rootCmd.PersistentFlags().StringSlice("nhl", []string{}, "NHL teams to watch")
	rootCmd.PersistentFlags().StringSlice("mlb", []string{}, "MLB teams to watch")
	rootCmd.PersistentFlags().StringSlice("cfl", []string{}, "CFL teams to watch")
	rootCmd.PersistentFlags().Bool("test-goals", false, "Enable or disable sending test goals every minute")
	rootCmd.PersistentFlags().Bool("web", false, "Start web interface mode")
	rootCmd.PersistentFlags().String("web-port", "8080", "Port for web interface")

	// Bind these flags to viper
	viper.BindPFlag("watch.nhl", rootCmd.PersistentFlags().Lookup("nhl"))
	viper.BindPFlag("watch.mlb", rootCmd.PersistentFlags().Lookup("mlb"))
	viper.BindPFlag("watch.cfl", rootCmd.PersistentFlags().Lookup("cfl"))
	viper.BindPFlag("test-goals", rootCmd.PersistentFlags().Lookup("test-goals"))
	viper.BindPFlag("web", rootCmd.PersistentFlags().Lookup("web"))
	viper.BindPFlag("web-port", rootCmd.PersistentFlags().Lookup("web-port"))

}

func main() {
	homeAssistantURL := os.Getenv("SUPERVISOR_API")
	fmt.Println(homeAssistantURL)
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runTickers() {
	var wg sync.WaitGroup
	tickers := []struct {
		duration time.Duration
		task     func()
	}{
		{1 * time.Minute, checkLeaguesForActiveGames},
		{1 * time.Second, watchActiveGames},
		{1 * time.Minute, sendTestGoal},
		{5 * time.Second, func() {
			if needRefresh {
				checkLeaguesForActiveGames()
				needRefresh = false
			}
		}},
	}

	for _, t := range tickers {
		wg.Add(1)
		go func(duration time.Duration, task func()) {
			defer wg.Done()
			ticker := time.NewTicker(duration)
			for range ticker.C {
				go task()
			}
		}(t.duration, t.task)
	}

	wg.Wait()
}

func runWebMode() {
	logger.Info("Starting Goalfeed in web mode")

	// Start the web server in a goroutine
	go webApi.StartWebServer(viper.GetString("web-port"))

	// Run the normal tickers
	runTickers()
}

func initialize() {
	logger.Info("Puck Drop! Initializing Goalfeed Process")

	leagueServices[models.LeagueIdNHL] = nhl.NHLService{Client: nhlClients.NHLApiClient{}}
	leagueServices[models.LeagueIdMLB] = mlb.MLBService{Client: mlbClients.MLBApiClient{}}
	leagueServices[models.LeagueIdCFL] = cfl.CFLService{Client: cflClients.CFLApiClient{}}
	leagueServices[models.LeagueIdNFL] = nfl.NFLService{Client: nflClients.NFLAPIClient{}}

	logger.Info("Initializing Active Games")
	checkLeaguesForActiveGames()
}

func checkLeaguesForActiveGames() {
	logger.Info("Updating Active Games")
	for _, service := range leagueServices {
		go checkForNewActiveGames(service)
	}
}

func checkForNewActiveGames(service leagues.ILeagueService) {
	logger.Info(fmt.Sprintf("Checking for active %s games", service.GetLeagueName()))
	gamesChan := make(chan []models.Game)
	go service.GetActiveGames(gamesChan)
	for _, game := range <-gamesChan {
		// Check if the home and away teams are being monitored
		if teamIsMonitoredByLeague(game.CurrentState.Home.Team.TeamCode, service.GetLeagueName()) ||
			teamIsMonitoredByLeague(game.CurrentState.Away.Team.TeamCode, service.GetLeagueName()) {
			if !gameIsMonitored(game) {
				logger.Info(fmt.Sprintf("Adding %s game (%s @ %s) to active monitored games", service.GetLeagueName(), game.CurrentState.Away.Team.TeamCode, game.CurrentState.Home.Team.TeamCode))
				memoryStore.SetGame(game)
				memoryStore.AppendActiveGame(game)
			}
		} else {
			logger.Info(fmt.Sprintf("Skipping %s game (%s @ %s) as teams are not being monitored", service.GetLeagueName(), game.CurrentState.Away.Team.TeamCode, game.CurrentState.Home.Team.TeamCode))
		}
	}
}

func gameIsMonitored(game models.Game) bool {
	for _, activeGameKey := range memoryStore.GetActiveGameKeys() {
		if activeGameKey == game.GetGameKey() {
			return true
		}
	}
	return false
}

func watchActiveGames() {
	for _, gameKey := range memoryStore.GetActiveGameKeys() {
		go checkGame(gameKey)
	}
}

func checkGame(gameKey string) {
	game, err := memoryStore.GetGameByGameKey(gameKey)
	if err != nil {
		logger.Error(err.Error())
		logger.Error(fmt.Sprintf("[%s] Game not found, skipping", gameKey))
		memoryStore.DeleteActiveGameKey(gameKey)
		needRefresh = true
		return
	}

	service := leagueServices[int(game.LeagueId)]
	logger.Info(fmt.Sprintf("[%s - %s %d @ %s %d] Checking", service.GetLeagueName(), game.CurrentState.Away.Team.TeamCode, game.CurrentState.Away.Score, game.CurrentState.Home.Team.TeamCode, game.CurrentState.Home.Score))
	game.IsFetching = true
	memoryStore.SetGame(game)

	updateChan := make(chan models.GameUpdate)
	eventChan := make(chan []models.Event)
	go service.GetGameUpdate(game, updateChan)
	update := <-updateChan
	go service.GetEvents(update, eventChan)
	go fireGoalEvents(eventChan, game)
	// Check for period changes and game state changes
	oldPeriod := game.CurrentState.Period
	oldStatus := game.CurrentState.Status

	game.CurrentState = update.NewState

	// Send period updates if period changed
	if oldPeriod != game.CurrentState.Period {
		if game.CurrentState.Period > oldPeriod {
			// Period started
			go homeassistant.SendPeriodUpdate(game, models.EventTypePeriodStart)
			logger.Info(fmt.Sprintf("[%s - %s @ %s] Period %d started", service.GetLeagueName(), game.CurrentState.Away.Team.TeamCode, game.CurrentState.Home.Team.TeamCode, game.CurrentState.Period))
		}
	}

	// Send game state updates
	if oldStatus != game.CurrentState.Status {
		switch game.CurrentState.Status {
		case models.StatusActive:
			if oldStatus == models.StatusUpcoming {
				go homeassistant.SendPeriodUpdate(game, models.EventTypeGameStart)
				logger.Info(fmt.Sprintf("[%s - %s @ %s] Game started", service.GetLeagueName(), game.CurrentState.Away.Team.TeamCode, game.CurrentState.Home.Team.TeamCode))
			}
		case models.StatusEnded:
			go homeassistant.SendPeriodUpdate(game, models.EventTypeGameEnd)
			logger.Info(fmt.Sprintf("[%s - %s @ %s] Game ended", service.GetLeagueName(), game.CurrentState.Away.Team.TeamCode, game.CurrentState.Home.Team.TeamCode))
		}
	}

	if game.CurrentState.Status == models.StatusEnded {
		memoryStore.DeleteActiveGame(game)
		memoryStore.DeleteActiveGameKey(game.GetGameKey()) // Ensure the game key is removed from active game keys
	} else {
		game.IsFetching = false
		memoryStore.SetGame(game)
		// Send enhanced game update to Home Assistant
		go homeassistant.SendGameUpdate(game)
		// Broadcast game update to web clients
		webApi.BroadcastGameUpdate(game)
	}
}

func fireGoalEvents(events chan []models.Event, game models.Game) {
	for _, event := range <-events {
		logger.Info(fmt.Sprintf("Event %s: %s", event.Type, event.Description))
		if teamIsMonitoredByLeague(event.TeamCode, leagueServices[int(game.LeagueId)].GetLeagueName()) {
			// Send enhanced event to Home Assistant
			go eventSender(event)
			// Broadcast event to web clients
			webApi.BroadcastEvent(event)
		}
	}
}
func teamIsMonitoredByLeague(teamCode, leagueName string) bool {
	// Convert leagueName to lowercase for consistency
	leagueName = strings.ToLower(leagueName)

	// Get the teams to watch for the given league from the configuration
	teamsToWatch := config.GetStringSlice("watch." + leagueName)

	// If "*" is in the watch list, monitor all teams for this league
	for _, team := range teamsToWatch {
		if team == "*" {
			return true
		}
	}

	// Check if the teamCode is in the list of teams to watch
	for _, team := range teamsToWatch {
		if strings.EqualFold(team, teamCode) {
			return true
		}
	}

	return false
}
func sendTestGoal() {
	if !viper.GetBool("test-goals") {
		logger.Info("Test goals are disabled. Skipping sending test goal.")
		return
	}
	logger.Info("Sending test goal")
	go eventSender(models.Event{
		TeamCode:     "TEST",
		TeamName:     "TEST",
		TeamHash:     "TESTTEST",
		LeagueId:     0,
		LeagueName:   "TEST",
		OpponentCode: "TEST",
		OpponentName: "TEST",
		OpponentHash: "TESTTEST",
	})
}
