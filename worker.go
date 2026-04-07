package main

import (
	"context"
	"log"
	"time"
)

func fetchLeaderboardWorker(ctx context.Context, service *PlayerService, ch chan<- []LeaderboardEntry) {
	for {
		leader, err := service.GetTopPlayers(context.Background(), 5, 0)
		if err != nil {
			log.Println(err.Error())
		} else {
			ch <- leader
			log.Println("New TOP 5")
		}
		select {
		case <-ctx.Done():
			log.Println("signal of stoping")
			return
		case <-time.After(10 * time.Second):
		}
	}
}
func updateCacheWorker(cache *LeaderboardCache, ch <-chan []LeaderboardEntry) {
	for newTop := range ch {
		cache.mu.Lock()
		cache.entries = newTop
		cache.mu.Unlock()
	}
}
