package main

import (
	"database/sql"
	"fmt"
	"strings"
)

type PlayerRepository struct {
	db *sql.DB
}

func (r *PlayerRepository) Create(p Player) error {
	_, err := r.db.Exec("INSERT INTO players (name, kills, deaths, matches) VALUES ($1, $2, $3, $4)",
		p.Name, p.Kills, p.Deaths, p.Matches)
	return err
}

func (r *PlayerRepository) GetAll() ([]Player, error) {
	rows, err := r.db.Query("SELECT name, kills, deaths, matches FROM players")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []Player

	for rows.Next() {
		var p Player
		rows.Scan(&p.Name, &p.Kills, &p.Deaths, &p.Matches)
		players = append(players, p)
	}
	return players, nil

}

func (r *PlayerRepository) Update(name string, update UpdatePlayer) error {
	var setClauses []string
	var args []any
	argPos := 1

	if update.Kills != nil {
		setClauses = append(setClauses, fmt.Sprintf("kills=$%d", argPos))
		args = append(args, *update.Kills)
		argPos++
	}

	if update.Deaths != nil {
		setClauses = append(setClauses, fmt.Sprintf("deaths=$%d", argPos))
		args = append(args, *update.Deaths)
		argPos++
	}

	if update.Matches != nil {
		setClauses = append(setClauses, fmt.Sprintf("matches=$%d", argPos))
		args = append(args, *update.Matches)
		argPos++
	}

	args = append(args, name)

	query := "UPDATE players SET " + strings.Join(setClauses, ", ") + " WHERE name=$" + fmt.Sprint(argPos)

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("player not found")
	}
	return nil
}

func (r *PlayerRepository) Delete(name string) error {
	result, err := r.db.Exec(
		"DELETE FROM player WHERE name = $1",
		name,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("player not found")
	}

	return nil
}
