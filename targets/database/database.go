package database

import (
	"database/sql"
	"goalfeed/models"
	"goalfeed/utils"
	"log"
	"os"
	"time"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/mattn/go-sqlite3"
)

func InitializeDatabase() error {
	db, err := getClient()
	if err != nil {
		return err
	}
	defer db.Close()

	// Create teams table if it doesn't exist
	teamsTable := `
	CREATE TABLE IF NOT EXISTS teams (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		team_code TEXT NOT NULL,
		team_name TEXT NOT NULL,
		league_id INTEGER NOT NULL,
		ext_id TEXT NOT NULL
	);
	`
	_, err = db.Exec(teamsTable)
	if err != nil {
		return err
	}

	// Create goals table if it doesn't exist
	goalsTable := `
	CREATE TABLE IF NOT EXISTS goals (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		team_id INTEGER NOT NULL,
		game_id INTEGER NOT NULL,
		timestamp INTEGER NOT NULL
	);
	`
	_, err = db.Exec(goalsTable)
	if err != nil {
		return err
	}

	return nil
}

func getClient() (*sql.DB, error) {
	dbPath := os.Getenv("DB_PATH") // Assuming you have a DB_PATH environment variable for SQLite file path
	// if dbPath isn't set, use the current folder and create goalfeed.db
	if dbPath == "" {
		dbPath = "./goalfeed.db"
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	return db, nil
}
func GetTeamByExtId(extId string, leagueId int) (models.Team, error) {
	db, _ := getClient()
	teamQ := sq.Select("*").From("teams").Where(sq.Eq{"ext_id": extId, "league_id": leagueId}).Limit(1)
	rows, err := teamQ.RunWith(db).Query()
	if err != nil {
		logger.Error(err)
		db.Close()
		return models.Team{}, err
	}
	var team models.Team

	if rows.Next() {
		err = rows.Scan(&team.ID, &team.TeamCode, &team.TeamName, &team.LeagueID, &team.ExtID)
		if err != nil {
			rows.Close()
			db.Close()
			return team, err
		}
	}
	rows.Close()
	db.Close()
	return team, nil
}

var logger = utils.GetLogger()

func GetOrCreateTeam(team models.Team) models.Team {
	existingTeam, e := GetTeamByExtId(team.ExtID, team.LeagueID)

	if e == nil && existingTeam.ID != 0 {
		if existingTeam.TeamName != team.TeamName {
			go UpdateTeam(team)
		}
		return existingTeam
	}
	newTeam, err := InsertTeam(team)
	if err != nil {
		logger.Error("Could not get or create team from DB", err)
		panic(err)
	}
	return newTeam

}

func InsertTeam(team models.Team) (models.Team, error) {
	db, _ := getClient()
	teamQ := sq.Insert("teams").
		Columns("team_code", "team_name", "league_id", "ext_id").
		Values(team.TeamCode, team.TeamName, team.LeagueID, team.ExtID)
	_, err := teamQ.RunWith(db).Query()
	if err != nil {
		//TODO, remove or rename
		db.Close()
		updatedTeam, err := UpdateTeam(team)
		if err != nil {
			db.Close()
			log.Fatal(err)
			panic(err)
		}
		return updatedTeam, nil
	}
	db.Close()
	return GetTeamByExtId(team.ExtID, team.LeagueID)
}
func InsertGoal(team models.Team) {
	db, _ := getClient()
	goalQ := sq.Insert("goals").
		Columns("team_id", "game_id", "timestamp").
		Values(team.ID, 0, int32(time.Now().Unix()))

	_, err := goalQ.RunWith(db).Query()
	if err != nil {
		logger.Error("Unable to log goal")
	}
	db.Close()
}

func UpdateTeam(team models.Team) (models.Team, error) {
	db, _ := getClient()
	teamQ := sq.Update("teams").
		Set("ext_id", team.ExtID).
		Set("team_name", team.TeamName).
		Where("league_id = ? AND team_code like ?", team.LeagueID, team.TeamCode)
	rows, err := teamQ.RunWith(db).Query()
	_ = rows
	if err != nil {
		db.Close()
		logger.Error(err)
	}
	db.Close()
	return GetTeamByExtId(team.ExtID, team.LeagueID)
}
