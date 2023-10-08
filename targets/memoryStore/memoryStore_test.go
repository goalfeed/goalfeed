package memoryStore

import (
	"goalfeed/models"
	"testing"
)

func TestActiveGameKeys(t *testing.T) {
	// Clear storage for a fresh start
	storage = make(map[string]string)

	// Test initial state
	if len(GetActiveGameKeys()) != 0 {
		t.Error("Expected no active game keys initially")
	}

	// Test setting active game keys
	SetActiveGameKeys([]string{"game1", "game2"})
	if len(GetActiveGameKeys()) != 2 {
		t.Error("Expected 2 active game keys after setting")
	}
}

func TestAppendAndDeleteActiveGame(t *testing.T) {
	// Clear storage for a fresh start
	storage = make(map[string]string)

	game := models.Game{
		GameCode: "game1",
		LeagueId: 1,
	}

	AppendActiveGame(game)
	if len(GetActiveGameKeys()) != 1 {
		t.Error("Expected 1 active game key after appending")
	}

	DeleteActiveGame(game)
	if len(GetActiveGameKeys()) != 0 {
		t.Error("Expected no active game keys after deletion")
	}
}

func TestGameByGameKey(t *testing.T) {
	// Clear storage for a fresh start
	storage = make(map[string]string)

	game := models.Game{
		GameCode: "game1",
	}

	SetGame(game)
	retrievedGame, err := GetGameByGameKey(game.GetGameKey())
	if err != nil {
		t.Errorf("Error retrieving game: %s", err)
	}

	if retrievedGame.GameCode != game.GameCode {
		t.Error("Retrieved game does not match original game")
	}
}

// Add more tests as needed for other functions
