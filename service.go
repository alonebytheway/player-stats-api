package main

import (
	"context"
	"errors"
	"sort"
)

var (
	ErrorNotFound     = errors.New("not found")
	ErrorBadRequest   = errors.New("bad request")
	ErrorInvalidStats = errors.New("invalid stats")
)

type Repository interface {
	Create(ctx context.Context, p Player) error
	GetByName(ctx context.Context, name string) (Player, error)
	GetAll(ctx context.Context) ([]Player, error)
	Update(ctx context.Context, name string, update UpdatePlayer) error
	Delete(ctx context.Context, name string) error
	GetTopPlayers(ctx context.Context, limit int, offset int) ([]Player, error)
	RecordDuel(ctx context.Context, winner string, loser string) error
}

func KD(p Player) float64 {
	if p.Deaths == 0 {
		return float64(p.Kills)
	}
	return float64(p.Kills) / float64(p.Deaths)
}

func AvgKills(p Player) float64 {
	if p.Matches == 0 {
		return 0
	}
	return float64(p.Kills) / float64(p.Matches)
}

func (s *PlayerService) CreatePlayer(ctx context.Context, p Player) error {
	if p.Name == "" {
		return ErrorBadRequest
	}
	if p.Kills < 0 || p.Deaths < 0 || p.Matches < 0 {
		return ErrorInvalidStats
	}
	return s.repo.Create(ctx, p)
}

func (s *PlayerService) GetAll(ctx context.Context) ([]Player, error) {
	return s.repo.GetAll(ctx)
}

func (s *PlayerService) Update(ctx context.Context, name string, update UpdatePlayer) error {
	if name == "" {
		return ErrorBadRequest
	}

	if update.Kills != nil && *update.Kills < 0 {
		return ErrorBadRequest
	}
	if update.Deaths != nil && *update.Deaths < 0 {
		return ErrorBadRequest
	}
	if update.Matches != nil && *update.Matches < 0 {
		return ErrorBadRequest
	}

	if update.Kills == nil && update.Deaths == nil && update.Matches == nil {
		return ErrorBadRequest
	}

	return s.repo.Update(ctx, name, update)
}

func (s *PlayerService) buildeLeaderboard(players []Player) ([]LeaderboardEntry, error) {

	players, err := s.repo.GetAll(context.Background())
	if err != nil {
		return nil, err
	}

	leaderboard := make([]LeaderboardEntry, 0, len(players))

	for _, p := range players {
		leaderboard = append(leaderboard, LeaderboardEntry{
			Name: p.Name,
			KD:   KD(p),
		})
	}

	sort.Slice(leaderboard, func(i, j int) bool {
		return leaderboard[i].KD > leaderboard[j].KD
	})

	return leaderboard, nil
}

func (s *PlayerService) DeletePlayer(ctx context.Context, name string) error {
	if name == "" {
		return ErrorBadRequest
	}

	err := s.repo.Delete(ctx, name)
	if err != nil {
		return err
	}

	return nil
}

func (s *PlayerService) GetPlayer(ctx context.Context, name string) (Player, error) {
	if name == "" {
		return Player{}, ErrorBadRequest
	}

	player, err := s.repo.GetByName(ctx, name)
	if err != nil {
		return Player{}, err
	}
	return player, nil
}

func (s *PlayerService) GetTopPlayers(ctx context.Context, limit int, offset int) ([]LeaderboardEntry, error) {
	if limit > 101 {
		limit = 100
	}

	players, err := s.repo.GetTopPlayers(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	leaderboard := make([]LeaderboardEntry, 0, len(players))
	for _, p := range players {
		leaderboard = append(leaderboard, LeaderboardEntry{
			Name: p.Name,
			KD:   KD(p),
		})
	}
	return leaderboard, nil
}

func (s *PlayerService) RecordDuel(ctx context.Context, winner string, loser string) error {
	return s.repo.RecordDuel(ctx, winner, loser)
}
