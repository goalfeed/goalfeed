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
	"goalfeed/targets/applog"
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

// TickerConfig holds configuration for a ticker
type TickerConfig struct {
	Duration time.Duration
	Task     func()
}

// TickerManager handles the complex logic of managing multiple tickers
type TickerManager struct {
	tickers []TickerConfig
	wg      sync.WaitGroup
}

// NewTickerManager creates a new TickerManager instance
func NewTickerManager() *TickerManager {
	return &TickerManager{
		tickers: []TickerConfig{
			{1 * time.Minute, checkLeaguesForActiveGames},
			{1 * time.Second, watchActiveGames},
			{1 * time.Minute, sendTestGoal},
			{10 * time.Minute, publishSchedules},
			{5 * time.Second, func() {
				if needRefresh {
					checkLeaguesForActiveGames()
					needRefresh = false
				}
			}},
		},
	}
}

// AddTicker adds a new ticker configuration
func (tm *TickerManager) AddTicker(duration time.Duration, task func()) {
	tm.tickers = append(tm.tickers, TickerConfig{
		Duration: duration,
		Task:     task,
	})
}

// StartTicker starts a single ticker with the given configuration
func (tm *TickerManager) StartTicker(config TickerConfig) {
	tm.wg.Add(1)
	go func(duration time.Duration, task func()) {
		defer tm.wg.Done()
		ticker := time.NewTicker(duration)
		defer ticker.Stop()
		for range ticker.C {
			go task()
		}
	}(config.Duration, config.Task)
}

// StartAllTickers starts all configured tickers
func (tm *TickerManager) StartAllTickers() {
	for _, config := range tm.tickers {
		tm.StartTicker(config)
	}
}

// WaitForCompletion waits for all tickers to complete
func (tm *TickerManager) WaitForCompletion() {
	tm.wg.Wait()
}

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
	tm := NewTickerManager()
	tm.StartAllTickers()
	tm.WaitForCompletion()
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

	// Publish baseline sensors for monitored teams at startup
	homeassistant.PublishBaselineForMonitoredTeams()

	// Start Fastcast listener for NFL if enabled
	nfl.StartNFLFastcast()
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
				// Immediately broadcast updated games list so UI gets the game right away
				webApi.BroadcastGamesList()
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
		return
	}

	service := leagueServices[int(game.LeagueId)]
	if service == nil {
		return
	}

	updateChan := make(chan models.GameUpdate)
	go service.GetGameUpdate(game, updateChan)
	gameUpdate := <-updateChan

	if gameUpdate.NewState.Period != gameUpdate.OldState.Period {
		logger.Info(fmt.Sprintf("Period change detected for %s game %s: %d -> %d", service.GetLeagueName(), game.GameCode, gameUpdate.OldState.Period, gameUpdate.NewState.Period))

		// Fire period change event
		event := models.Event{
			TeamCode:    "",
			TeamName:    "",
			LeagueId:    int(game.LeagueId),
			LeagueName:  service.GetLeagueName(),
			Type:        models.EventTypePeriodStart,
			Description: fmt.Sprintf("Period %d started", gameUpdate.NewState.Period),
		}
		eventSender(event)
	}

	// Update the game with new state
	updatedGame := game
	updatedGame.CurrentState = gameUpdate.NewState
	memoryStore.SetGame(updatedGame)
}

func fireGoalEvents(events chan []models.Event, game models.Game) {
	for _, event := range <-events {
		logger.Info(fmt.Sprintf("Event %s: %s", event.Type, event.Description))
		if teamIsMonitoredByLeague(event.TeamCode, leagueServices[int(game.LeagueId)].GetLeagueName()) {
			// Send enhanced event to Home Assistant
			go eventSender(event)
			// Append to app log
			go applog.AppendEvent(event)
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

func publishSchedules() {
	logger.Info("Publishing schedule sensors")
	leagueConfigs := []struct {
		id   models.League
		name string
	}{
		{models.LeagueIdNHL, "nhl"},
		{models.LeagueIdMLB, "mlb"},
		{models.LeagueIdCFL, "cfl"},
		{models.LeagueIdNFL, "nfl"},
	}

	for _, lc := range leagueConfigs {
		teams := config.GetStringSlice("watch." + lc.name)
		if len(teams) == 0 {
			continue
		}
		svc := leagueServices[int(lc.id)]
		if svc == nil {
			continue
		}
		ch := make(chan []models.Game)
		go svc.GetUpcomingGames(ch)
		upcoming := <-ch
		for _, g := range upcoming {
			if teamIsMonitoredByLeague(g.CurrentState.Home.Team.TeamCode, lc.name) || teamIsMonitoredByLeague(g.CurrentState.Away.Team.TeamCode, lc.name) {
				if g.CurrentState.Status == models.StatusUpcoming || (!g.GameDetails.GameDate.IsZero() && g.GameDetails.GameDate.After(time.Now().Add(-1*time.Hour))) {
					homeassistant.PublishScheduleSensorsForGame(g)
				}
			}
		}
	}
}
