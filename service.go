package main

import (
	"fmt"
	"sort"
)

var (
	ErrorNotFound   = fmt.Errorf("not found")
	ErrorBadRequest = fmt.Errorf("bad request")
)

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

func (s *PlayerService) CreatePlayer(p Player) error {
	if p.Name == "" {
		return fmt.Errorf("name is required")
	}
	if p.Kills < 0 || p.Deaths < 0 || p.Matches < 0 {
		return fmt.Errorf("invalid stats")
	}
	return s.repo.Create(p)
}

func (s *PlayerService) GetAll() ([]Player, error) {
	return s.repo.GetAll()
}

func (s *PlayerService) Update(name string, update UpdatePlayer) error {
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

	err := s.repo.Update(name, update)
	if err != nil {
		if err.Error() == "player not found" {
			return ErrorNotFound
		}
		return err
	}
	return nil
}

func (s *PlayerService) GetLeaderboard() ([]LeaderboardEntry, error) {

	players, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	var leaderboard []LeaderboardEntry

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

func (s *PlayerService) DeletePlayer(name string) error {
	if name == "" {
		return ErrorBadRequest
	}

	err := s.repo.Delete(name)
	if err != nil {
		if err.Error() == "player not found" {
			return ErrorNotFound
		}
		return err
	}

	return nil
}

func (s *PlayerService) GetPlayer(name string) (Player, error) {
	if name == "" {
		return Player{}, ErrorBadRequest
	}
	player, err := s.repo.GetByName(name)
	if err != nil {
		if err == ErrorNotFound {
			return Player{}, ErrorNotFound
		}
		return Player{}, err
	}
	return player, nil
}
