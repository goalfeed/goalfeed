package leagues

import (
	"goalfeed/models"
)

type ILeagueService interface {
	GetLeagueName() string
	GetActiveGames(ret chan []models.Game)
	GetUpcomingGames(ret chan []models.Game)
	GetGamesByDate(date string, ret chan []models.Game)
	GetGameUpdate(game models.Game, ret chan models.GameUpdate)
	GetEvents(update models.GameUpdate, ret chan []models.Event)
}
