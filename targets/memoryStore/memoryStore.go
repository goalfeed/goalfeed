package memoryStore

import (
	"encoding/json"
	"fmt"
	"goalfeed/models"
	"goalfeed/utils"
	"sync"
)

var logger = utils.GetLogger()

// In-memory storage
var storage = make(map[string]string)
var storageMutex = &sync.RWMutex{}

const ACTIVE_GAME_CODES_KEY = "GoalfeedActiveGamesv1"

func GetActiveGameKeys() []string {
	storageMutex.RLock()
	gamesJSON, exists := storage[ACTIVE_GAME_CODES_KEY]
	storageMutex.RUnlock()

	if !exists {
		return []string{}
	}

	var activeGameKeys []string
	json.Unmarshal([]byte(gamesJSON), &activeGameKeys)
	logger.Debug(gamesJSON)
	return activeGameKeys
}

func SetActiveGameKeys(gameCodes []string) {
	gamesByte, _ := json.Marshal(gameCodes)
	storageMutex.Lock()
	storage[ACTIVE_GAME_CODES_KEY] = string(gamesByte)
	storageMutex.Unlock()
}

func AppendActiveGame(game models.Game) {
	activeGameKeys := GetActiveGameKeys()
	SetGame(game)
	SetActiveGameKeys(append(activeGameKeys, game.GetGameKey()))
}

func DeleteActiveGame(game models.Game) {
	DeleteActiveGameKey(game.GetGameKey())
}

func DeleteActiveGameKey(gameKey string) {
	activeGameKeys := GetActiveGameKeys()
	for i, gameCode := range activeGameKeys {
		if gameCode == gameKey {
			activeGameKeys = append(activeGameKeys[:i], activeGameKeys[i+1:]...)
			break
		}
	}
	SetActiveGameKeys(activeGameKeys)
}

func GetGameByGameKey(gameCode string) (models.Game, error) {
	storageMutex.RLock()
	gameJSON, exists := storage[gameCode]
	storageMutex.RUnlock()

	if !exists {
		return models.Game{}, fmt.Errorf("Game not found")
	}

	var game models.Game
	json.Unmarshal([]byte(gameJSON), &game)
	logger.Debug(gameJSON)
	return game, nil
}

func SetGame(game models.Game) {
	logger.Debug(fmt.Sprintf("writing to key %s", game.GetGameKey()))
	gameByte, err := json.Marshal(game)
	if err != nil {
		panic(err)
	}
	logger.Debug(fmt.Sprintf("writing %s to key %s", string(gameByte), game.GetGameKey()))

	storageMutex.Lock()
	storage[game.GetGameKey()] = string(gameByte)
	storageMutex.Unlock()
}

func GetAllGames() []models.Game {
	activeGameKeys := GetActiveGameKeys()
	var games []models.Game

	for _, gameKey := range activeGameKeys {
		if game, err := GetGameByGameKey(gameKey); err == nil {
			games = append(games, game)
		}
	}

	return games
}
