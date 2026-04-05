package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type PlayerRepository struct {
	db *sql.DB
}

func (r *PlayerRepository) Create(ctx context.Context, p Player) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO players (name, kills, deaths, matches) VALUES ($1, $2, $3, $4)",
		p.Name, p.Kills, p.Deaths, p.Matches)
	return err
}

func (r *PlayerRepository) GetByName(name string) (Player, error) {
	row := r.db.QueryRow("SELECT name, kills, deaths, matches FROM players WHERE name = $1", name)
	var p Player
	err := row.Scan(&p.Name, &p.Kills, &p.Deaths, &p.Matches)
	if err != nil {
		return Player{}, err
	}
	return p, nil
}

func (r *PlayerRepository) GetAll(ctx context.Context) ([]Player, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT name, kills, deaths, matches FROM players")
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
	if len(setClauses) == 0 {
		return fmt.Errorf("no fields to update")
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
		return ErrorNotFound
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
		return ErrorNotFound
	}

	return nil
}

func (r *PlayerRepository) GetTopPlayers(ctx context.Context, limit int, offset int) ([]Player, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT name, kills, deaths, matches FROM players ORDER BY(CASE WHEN deaths = 0 THEN kills ELSE(CAST(kills AS FLOAT)/ deaths) END) DESC LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	players := make([]Player, 0, limit)

	for rows.Next() {
		var p Player
		if err := rows.Scan(&p.Name, &p.Kills, &p.Deaths, &p.Matches); err != nil {
			return nil, err
		}
		players = append(players, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return players, nil
}

func (r *PlayerRepository) RecordDuel(ctx context.Context, winner string, loser string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "UPDATE players SET kills = kills + 1, matches = matches + 1 WHERE name = $1", winner)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "UPDATE players SET deaths = deaths + 1, matches = matches + 1 WHERE name = $1", loser)
	if err != nil {
		return err
	}
	return tx.Commit()
}
